package request

import (
	"net/http"
	"strings"

	"k8s.io/apimachinery/pkg/api/validation/path"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

type RequestInfoResolver interface {
	NewRequestInfo(req *http.Request) (*RequestInfo, error)
}

type RequestInfo struct {
	// Path is the URL path of the request
	Path string
	// Verb is the kube verb associated with the request for API requests, not the http verb.  This includes things like list and watch.
	// for non-resource requests, this is the lowercase http verb
	Verb string

	APIPrefix  string
	APIGroup   string
	APIVersion string
	Namespace  string
	Cluster    string
	Project    string
	Level      string
	// Resource is the name of the resource being requested.  This is not the kind.  For example: pods
	Resource string
	// Subresource is the name of the subresource being requested.  This is a different resource, scoped to the parent resource, but it may have a different kind.
	// For instance, /pods has the resource "pods" and the kind "Pod", while /pods/foo/status has the resource "pods", the sub resource "status", and the kind "Pod"
	// (because status operates on pods). The binding resource for a pod though may be /pods/foo/binding, which has resource "pods", subresource "binding", and kind "Binding".
	Subresource string
	// Name is empty for some verbs, but if the request directly indicates a name (not in body content) then this field is filled in.
	Name string
	// Parts are the path parts for the request, always starting with /{resource}/{name}
	Parts []string
}

type RequestInfoFactory struct {
	APIPrefixes          sets.String // without leading and trailing slashes
	GrouplessAPIPrefixes sets.String // without leading and trailing slashes
}

// Resource paths
// /{product-prefix}/{api-group}/{version}/projects/{project}/{resource}
// /{product-prefix}/{api-group}/{version}/projects/{project}/{resource}/{resourceName}
// /{product-prefix}/{api-group}/{version}/projects/{project}/clusters/{cluster}/namespaces/{namespace}/{resource}
// /{product-prefix}/{api-group}/{version}/projects/{project}/clusters/{cluster}/namespaces/{namespace}/{resource}/{resourceName}
// /{product-prefix}/{api-group}/{version}/clusters/{cluster}/{resource}
// /{product-prefix}/{api-group}/{version}/clusters/{cluster}/{resource}/{resourceName}
// /{product-prefix}/{api-group}/{version}/{resource}
// /{product-prefix}/{api-group}/{version}/{resource}/{resourceName}

func (r *RequestInfoFactory) NewRequestInfo(req *http.Request) (*RequestInfo, error) {
	requestInfo := RequestInfo{
		Path: req.URL.Path,
		Verb: strings.ToLower(req.Method),
	}

	switch req.Method {
	case "POST":
		requestInfo.Verb = "create"
	case "GET", "HEAD":
		requestInfo.Verb = "get"
	case "PUT":
		requestInfo.Verb = "update"
	case "PATCH":
		requestInfo.Verb = "patch"
	case "DELETE":
		requestInfo.Verb = "delete"
	default:
		requestInfo.Verb = ""
	}

	currentParts := splitPath(req.URL.Path)
	if !r.APIPrefixes.Has(currentParts[0]) {
		return &requestInfo, nil
	}
	// find product-prefix
	requestInfo.APIPrefix = currentParts[0]
	currentParts = currentParts[1:]

	// find api-group
	// TODO core group is empty
	requestInfo.APIGroup = currentParts[0]
	currentParts = currentParts[1:]

	// find version
	requestInfo.APIVersion = currentParts[0]
	currentParts = currentParts[1:]

	// find project, cluster, namespace
	// case1: /projects/{project}/{resource}
	// case2: /projects/{project}/clusters/{cluster}/namespaces/{namespace}/{resource}
	if currentParts[0] == "projects" {
		if len(currentParts) > 1 {
			requestInfo.Project = currentParts[1]
			requestInfo.Level = "project"

			if len(currentParts) > 2 {
				currentParts = currentParts[2:]

				// case1: /{resource}
				// case2: /clusters/{cluster}/namespaces/{namespace}/{resource}
				if currentParts[0] == "clusters" {
					if len(currentParts) > 1 {
						requestInfo.Cluster = currentParts[1]

						if len(currentParts) > 2 {
							currentParts = currentParts[2:]

							// case1: /namespaces/{namespace}/{resource}
							if currentParts[0] == "namespaces" {
								requestInfo.Namespace = currentParts[1]
								requestInfo.Level = "namespace"

								if len(currentParts) > 2 {
									currentParts = currentParts[2:]
								}
							}
						}
					}
				}
			}
		}
		// URL forms: /clusters/{cluster}/{resource}
	} else if currentParts[0] == "clusters" {
		// cluster
		if len(currentParts) > 1 {
			requestInfo.Cluster = currentParts[1]
			requestInfo.Project = ""
			requestInfo.Namespace = ""
			requestInfo.Level = "cluster"

			if len(currentParts) > 2 {
				currentParts = currentParts[2:]
			}
		}

	} else {
		// platform
		requestInfo.Namespace = metav1.NamespaceNone
		requestInfo.Project = ""
		requestInfo.Cluster = ""
		requestInfo.Level = "platform"
	}

	// parsing successful, so we now know the proper value for .Parts
	requestInfo.Parts = currentParts

	// parts look like: resource/resourceName/subresource/other/stuff/we/don't/interpret
	switch {
	case len(requestInfo.Parts) >= 3:
		requestInfo.Subresource = requestInfo.Parts[2]
		fallthrough
	case len(requestInfo.Parts) >= 2:
		requestInfo.Name = requestInfo.Parts[1]
		fallthrough
	case len(requestInfo.Parts) >= 1:
		requestInfo.Resource = requestInfo.Parts[0]
	}

	// if there's no name on the request and we thought it was a get before, then the actual verb is a list or a watch
	if len(requestInfo.Name) == 0 && requestInfo.Verb == "get" {
		opts := metainternalversion.ListOptions{}
		if opts.Watch {
			requestInfo.Verb = "watch"
		} else {
			requestInfo.Verb = "list"
		}

		if opts.FieldSelector != nil {
			if name, ok := opts.FieldSelector.RequiresExactMatch("metadata.name"); ok {
				if len(path.IsValidPathSegmentName(name)) == 0 {
					requestInfo.Name = name
				}
			}
		}
	}
	// if there's no name on the request and we thought it was a delete before, then the actual verb is deletecollection
	if len(requestInfo.Name) == 0 && requestInfo.Verb == "delete" {
		requestInfo.Verb = "deletecollection"
	}

	return &requestInfo, nil
}

// splitPath returns the segments for a URL path.
func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}
