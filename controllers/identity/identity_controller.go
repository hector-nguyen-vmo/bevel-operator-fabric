package identity

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/protoutil"
	hlfv1alpha1 "github.com/kfsoftware/hlf-operator/api/hlf.kungfusoftware.es/v1alpha1"
	"github.com/kfsoftware/hlf-operator/controllers/certs"
	"github.com/kfsoftware/hlf-operator/controllers/utils"
	"github.com/kfsoftware/hlf-operator/internal/github.com/hyperledger/fabric-ca/api"
	"github.com/kfsoftware/hlf-operator/pkg/status"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

// FabricIdentityReconciler reconciles a FabricIdentity object
type FabricIdentityReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Config *rest.Config
}

const identityFinalizer = "finalizer.identity.hlf.kungfusoftware.es"

func (r *FabricIdentityReconciler) finalizeMainChannel(reqLogger logr.Logger, m *hlfv1alpha1.FabricIdentity) error {
	ns := m.Namespace
	if ns == "" {
		ns = "default"
	}
	reqLogger.Info("Successfully finalized identity")

	return nil
}

func (r *FabricIdentityReconciler) addFinalizer(reqLogger logr.Logger, m *hlfv1alpha1.FabricIdentity) error {
	reqLogger.Info("Adding Finalizer for the MainChannel")
	controllerutil.AddFinalizer(m, identityFinalizer)

	// Update CR
	err := r.Update(context.TODO(), m)
	if err != nil {
		reqLogger.Error(err, "Failed to update MainChannel with finalizer")
		return err
	}
	return nil
}

// +kubebuilder:rbac:groups=hlf.kungfusoftware.es,resources=fabricidentities,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hlf.kungfusoftware.es,resources=fabricidentities/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hlf.kungfusoftware.es,resources=fabricidentities/finalizers,verbs=get;update;patch
func (r *FabricIdentityReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := r.Log.WithValues("hlf", req.NamespacedName)
	fabricIdentity := &hlfv1alpha1.FabricIdentity{}

	err := r.Get(ctx, req.NamespacedName, fabricIdentity)
	if err != nil {
		log.Debugf("Error getting the object %s error=%v", req.NamespacedName, err)
		if apierrors.IsNotFound(err) {
			reqLogger.Info("MainChannel resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		reqLogger.Error(err, "Failed to get MainChannel.")
		return ctrl.Result{}, err
	}
	markedToBeDeleted := fabricIdentity.GetDeletionTimestamp() != nil
	if markedToBeDeleted {
		if utils.Contains(fabricIdentity.GetFinalizers(), identityFinalizer) {
			if err := r.finalizeMainChannel(reqLogger, fabricIdentity); err != nil {
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(fabricIdentity, identityFinalizer)
			err := r.Update(ctx, fabricIdentity)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}
	if !utils.Contains(fabricIdentity.GetFinalizers(), identityFinalizer) {
		if err := r.addFinalizer(reqLogger, fabricIdentity); err != nil {
			return ctrl.Result{}, err
		}
	}
	clientSet, err := utils.GetClientKubeWithConf(r.Config)
	if err != nil {
		r.setConditionStatus(ctx, fabricIdentity, hlfv1alpha1.FailedStatus, false, err, false)
		return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricIdentity)
	}
	tlsCert, err := base64.StdEncoding.DecodeString(fabricIdentity.Spec.Catls.Cacert)
	if err != nil {
		r.setConditionStatus(ctx, fabricIdentity, hlfv1alpha1.FailedStatus, false, err, false)
		return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricIdentity)
	}
	x509Cert, pk, rootCert, err := certs.EnrollUser(certs.EnrollUserRequest{
		TLSCert:    string(tlsCert),
		URL:        fmt.Sprintf("https://%s:%d", fabricIdentity.Spec.Cahost, fabricIdentity.Spec.Caport),
		Name:       fabricIdentity.Spec.Caname,
		MSPID:      fabricIdentity.Spec.MSPID,
		User:       fabricIdentity.Spec.Enrollid,
		Secret:     fabricIdentity.Spec.Enrollsecret,
		Hosts:      []string{},
		Attributes: []*api.AttributeRequest{},
	})
	if err != nil {
		r.setConditionStatus(ctx, fabricIdentity, hlfv1alpha1.FailedStatus, false, err, false)
		return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricIdentity)
	}
	pkBytes, err := utils.EncodePrivateKey(pk)
	if err != nil {
		r.setConditionStatus(ctx, fabricIdentity, hlfv1alpha1.FailedStatus, false, err, false)
		return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricIdentity)
	}
	// get secret if exists
	secretExists := true
	secret, err := clientSet.CoreV1().Secrets(fabricIdentity.Namespace).Get(ctx, fabricIdentity.Name, v1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			secretExists = false
		} else {
			r.setConditionStatus(ctx, fabricIdentity, hlfv1alpha1.FailedStatus, false, err, false)
			return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricIdentity)
		}
	}
	if secretExists {
		secret.Data = map[string][]byte{
			"cert.pem": utils.EncodeX509Certificate(x509Cert),
			"key.pem":  pkBytes,
			"root.pem": utils.EncodeX509Certificate(rootCert),
		}
		if err := controllerutil.SetControllerReference(fabricIdentity, secret, r.Scheme); err != nil {
			r.setConditionStatus(ctx, fabricIdentity, hlfv1alpha1.FailedStatus, false, err, false)
			return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricIdentity)
		}
		if err := r.Update(ctx, secret); err != nil {
			r.setConditionStatus(ctx, fabricIdentity, hlfv1alpha1.FailedStatus, false, err, false)
			return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricIdentity)
		}
	} else {
		secret = &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      fabricIdentity.Name,
				Namespace: fabricIdentity.Namespace,
			},
			Data: map[string][]byte{
				"cert.pem": utils.EncodeX509Certificate(x509Cert),
				"key.pem":  pkBytes,
				"root.pem": utils.EncodeX509Certificate(rootCert),
			},
		}
		if err := controllerutil.SetControllerReference(fabricIdentity, secret, r.Scheme); err != nil {
			r.setConditionStatus(ctx, fabricIdentity, hlfv1alpha1.FailedStatus, false, err, false)
			return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricIdentity)
		}
		if err := r.Create(ctx, secret); err != nil {
			r.setConditionStatus(ctx, fabricIdentity, hlfv1alpha1.FailedStatus, false, err, false)
			return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricIdentity)
		}
	}
	fabricIdentity.Status.Status = hlfv1alpha1.RunningStatus
	fabricIdentity.Status.Message = "Identity Setup"
	fabricIdentity.Status.Conditions.SetCondition(status.Condition{
		Type:               "CREATED",
		Status:             "True",
		LastTransitionTime: v1.Time{},
	})
	if err := r.Status().Update(ctx, fabricIdentity); err != nil {
		r.setConditionStatus(ctx, fabricIdentity, hlfv1alpha1.FailedStatus, false, err, false)
		return r.updateCRStatusOrFailReconcile(ctx, r.Log, fabricIdentity)
	}
	return ctrl.Result{}, nil
}

var (
	ErrClientK8s = errors.New("k8sAPIClientError")
)

func (r *FabricIdentityReconciler) updateCRStatusOrFailReconcile(ctx context.Context, log logr.Logger, p *hlfv1alpha1.FabricIdentity) (
	reconcile.Result, error) {
	if err := r.Status().Update(ctx, p); err != nil {
		log.Error(err, fmt.Sprintf("%v failed to update the application status", ErrClientK8s))
		return reconcile.Result{}, err
	}
	return reconcile.Result{
		Requeue:      true,
		RequeueAfter: 10 * time.Second,
	}, nil
}

func (r *FabricIdentityReconciler) setConditionStatus(ctx context.Context, p *hlfv1alpha1.FabricIdentity, conditionType hlfv1alpha1.DeploymentStatus, statusFlag bool, err error, statusUnknown bool) (update bool) {
	statusStr := func() corev1.ConditionStatus {
		if statusUnknown {
			return corev1.ConditionUnknown
		}
		if statusFlag {
			return corev1.ConditionTrue
		} else {
			return corev1.ConditionFalse
		}
	}
	if p.Status.Status != conditionType {
		depCopy := client.MergeFrom(p.DeepCopy())
		p.Status.Status = conditionType
		err = r.Status().Patch(ctx, p, depCopy)
		if err != nil {
			log.Warnf("Failed to update status to %s: %v", conditionType, err)
		}
	}
	if err != nil {
		p.Status.Message = err.Error()
	}
	condition := func() status.Condition {
		if err != nil {
			return status.Condition{
				Type:    status.ConditionType(conditionType),
				Status:  statusStr(),
				Reason:  status.ConditionReason(err.Error()),
				Message: err.Error(),
			}
		}
		return status.Condition{
			Type:   status.ConditionType(conditionType),
			Status: statusStr(),
		}
	}
	return p.Status.Conditions.SetCondition(condition())
}

func (r *FabricIdentityReconciler) SetupWithManager(mgr ctrl.Manager) error {
	managedBy := ctrl.NewControllerManagedBy(mgr)
	return managedBy.
		For(&hlfv1alpha1.FabricIdentity{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}

type identity struct {
	Cert Pem `json:"cert"`
	Key  Pem `json:"key"`
}

type Pem struct {
	Pem string
}

func CreateConfigUpdateEnvelope(channelID string, configUpdate *common.ConfigUpdate) ([]byte, error) {
	configUpdate.ChannelId = channelID
	configUpdateData, err := proto.Marshal(configUpdate)
	if err != nil {
		return nil, err
	}
	configUpdateEnvelope := &common.ConfigUpdateEnvelope{}
	configUpdateEnvelope.ConfigUpdate = configUpdateData
	envelope, err := protoutil.CreateSignedEnvelope(common.HeaderType_CONFIG_UPDATE, channelID, nil, configUpdateEnvelope, 0, 0)
	if err != nil {
		return nil, err
	}
	envelopeData, err := proto.Marshal(envelope)
	if err != nil {
		return nil, err
	}
	return envelopeData, nil
}
