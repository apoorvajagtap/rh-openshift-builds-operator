package controller

//+kubebuilder:rbac:groups=operator.openshift.io,resources=openshiftbuilds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.openshift.io,resources=openshiftbuilds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.openshift.io,resources=openshiftbuilds/finalizers,verbs=update

//+kubebuilder:rbac:groups=operator.shipwright.io,resources=shipwrightbuilds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.shipwright.io,resources=shipwrightbuilds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.shipwright.io,resources=shipwrightbuilds/finalizers,verbs=update
//+kubebuilder:rbac:groups=storage.k8s.io,resources=csidrivers,resourceNames=csi.sharedresource.openshift.io,verbs=get;list;delete;patch
//+kubebuilder:rbac:groups=core,resources=services;pods;configmaps;secrets;events;namespaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=daemonsets,resourceNames=shared-resource-csi-driver-node,verbs=get;list;delete;patch
//+kubebuilder:rbac:groups=apps,resources=deployments,resourceNames=shared-resource-csi-driver-webhook,verbs=get;list;delete;patch
//+kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,resourceNames=shared-resource-csi-driver-pdb,verbs=get;list;delete;patch
//+kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations,resourceNames=validation.webhook.csidriversharedresource,verbs=get;list;delete;patch
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitor,resourceNames=shared-resource-csi-driver-node-monitor,verbs=get;list;delete;patch
//+kubebuilder:rbac:groups=sharedresource.openshift.io,resources=sharedconfigmaps;sharedsecrets,verbs=get;list;delete;create;update;patch
