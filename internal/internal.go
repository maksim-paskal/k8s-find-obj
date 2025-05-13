package internal

import (
	"context"
	"log/slog"
	"regexp"
	"slices"
	"strings"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func NewApplication() *Application {
	return &Application{
		KubernetesObjects: make([]KubernetesObject, 0),
		ShowTails:         10,
	}
}

type Application struct {
	clientset         *kubernetes.Clientset
	Kubeconfig        string
	WhereToSearch     string
	WhatToSearch      string
	whatToSearchRe    *regexp.Regexp
	Namespace         string
	KubernetesObjects []KubernetesObject
	ShowTails         int
	Except            string
	exceptRe          *regexp.Regexp
}

type KubernetesObject struct {
	Kind      string
	Name      string
	Namespace string
	Object    string
}

func (a *Application) Validate() error {
	if a.Kubeconfig == "" {
		return errors.New("kubeconfig is required")
	}

	if a.WhereToSearch == "" {
		return errors.New("where-to-search is required")
	}

	if a.WhatToSearch == "" {
		return errors.New("what-to-search is required")
	}

	return nil

}

func (a *Application) Init(ctx context.Context) error {
	whatToSearchRe, err := regexp.Compile(a.WhatToSearch)
	if err != nil {
		return errors.Wrap(err, "error in regexp.Compile "+a.WhatToSearch)
	}

	if a.Except != "" {
		exceptRe, err := regexp.Compile(a.Except)
		if err != nil {
			return errors.Wrap(err, "error in regexp.Compile "+a.Except)
		}

		a.exceptRe = exceptRe
	}

	restconfig, err := clientcmd.BuildConfigFromFlags("", a.Kubeconfig)
	if err != nil {
		return errors.Wrap(err, "error in clientcmd.BuildConfigFromFlags")
	}

	clientset, err := kubernetes.NewForConfig(restconfig)
	if err != nil {
		return errors.Wrap(err, "error in kubernetes.NewForConfig")
	}

	a.whatToSearchRe = whatToSearchRe
	a.clientset = clientset

	return nil
}

func (a *Application) isInWhere(obj string) bool {
	objs := strings.Split(strings.ToLower(a.WhereToSearch), ",")

	return slices.Contains(objs, strings.ToLower(obj))
}

func (a *Application) getPods(ctx context.Context) error {
	const typeOf = "Pods"

	if !a.isInWhere(typeOf) {
		return nil
	}

	slog.Info("Getting " + typeOf + " ...")

	objects, err := a.clientset.CoreV1().Pods(a.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "error in "+typeOf)
	}

	for _, object := range objects.Items {
		a.KubernetesObjects = append(a.KubernetesObjects, KubernetesObject{
			Kind:      typeOf,
			Name:      object.Name,
			Namespace: object.Namespace,
			Object:    object.String(),
		})
	}

	return nil
}

func (a *Application) getConfigmaps(ctx context.Context) error {
	const typeOf = "ConfigMaps"

	if !a.isInWhere(typeOf) {
		return nil
	}

	slog.Info("Getting " + typeOf + " ...")

	objects, err := a.clientset.CoreV1().ConfigMaps(a.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "error in "+typeOf)
	}

	for _, object := range objects.Items {
		a.KubernetesObjects = append(a.KubernetesObjects, KubernetesObject{
			Kind:      typeOf,
			Name:      object.Name,
			Namespace: object.Namespace,
			Object:    object.String(),
		})
	}

	return nil
}

func (a *Application) getDeployments(ctx context.Context) error {
	const typeOf = "Deployments"

	if !a.isInWhere(typeOf) {
		return nil
	}

	slog.Info("Getting " + typeOf + " ...")

	objects, err := a.clientset.AppsV1().Deployments(a.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "error in "+typeOf)
	}

	for _, object := range objects.Items {
		a.KubernetesObjects = append(a.KubernetesObjects, KubernetesObject{
			Kind:      typeOf,
			Name:      object.Name,
			Namespace: object.Namespace,
			Object:    object.String(),
		})
	}

	return nil
}

func (a *Application) getStatefulSets(ctx context.Context) error {
	const typeOf = "StatefulSets"

	if !a.isInWhere(typeOf) {
		return nil
	}

	slog.Info("Getting " + typeOf + " ...")

	objects, err := a.clientset.AppsV1().StatefulSets(a.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "error in "+typeOf)
	}

	for _, object := range objects.Items {
		a.KubernetesObjects = append(a.KubernetesObjects, KubernetesObject{
			Kind:      typeOf,
			Name:      object.Name,
			Namespace: object.Namespace,
			Object:    object.String(),
		})
	}

	return nil
}

func (a *Application) getCronJobs(ctx context.Context) error {
	const typeOf = "CronJobs"

	if !a.isInWhere(typeOf) {
		return nil
	}

	slog.Info("Getting " + typeOf + " ...")

	objects, err := a.clientset.BatchV1().CronJobs(a.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "error in "+typeOf)
	}

	for _, object := range objects.Items {
		a.KubernetesObjects = append(a.KubernetesObjects, KubernetesObject{
			Kind:      typeOf,
			Name:      object.Name,
			Namespace: object.Namespace,
			Object:    object.String(),
		})
	}

	return nil
}

func (a *Application) search() {
	for _, obj := range a.KubernetesObjects {
		slog := slog.With(
			"kind", obj.Kind,
			"name", obj.Name,
			"namespace", obj.Namespace,
		)

		if a.exceptRe != nil && a.exceptRe.MatchString(obj.Namespace+"/"+obj.Name) {
			slog.Debug("ignored")
			continue
		}

		locs := a.whatToSearchRe.FindAllStringIndex(strings.ToLower(obj.Object), -1)

		if locs == nil {
			continue
		}

		for _, loc := range locs {
			start := loc[0] - a.ShowTails
			end := loc[1] + a.ShowTails

			if start < 0 {
				start = 0
			}

			if max := len(obj.Object); end > max {
				end = max
			}

			text := obj.Object[start:end]
			text = strings.ReplaceAll(text, "\n", " ")

			slog.Info(text)
		}
	}
}

type searchFunc func(context.Context) error

func (a *Application) Run(ctx context.Context) error {
	searchFuncs := []searchFunc{
		a.getPods,
		a.getConfigmaps,
		a.getDeployments,
		a.getStatefulSets,
		a.getCronJobs,
	}

	for _, f := range searchFuncs {
		if err := f(ctx); err != nil {
			return err
		}
	}

	a.search()

	return nil
}
