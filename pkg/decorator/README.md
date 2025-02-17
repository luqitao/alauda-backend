# gomod.alauda.cn/alauda-backend/pkg/decorator

This package offers a few `structs` and functions to facilitate creationg of APIs.

## Client

Automatically initiates `k8s.io/client-go` clients injecting directly into the `request.Request.Context()`

**Constructor:**: `NewClient(srv server.Server)`

**Methods:**:

### Methods
#### InsecureFilter

a `restful.FilterFunction` (middleware) to create a `kubernetes.Interface` instance using the pod's service account (or kubeconfig data if provided). Will inject the client inside the `request.Request.Context()` object and can be retrieved using the `context.Client` method.

```
InsecureFilter(req *restful.Request, res *restful.Response, chain *restful.FilterChain) 
```

#### Secure

a `restful.FilterFunction` (middleware) to create a `kubernetes.Interface` instance using provided `Bearer token`. Will inject the client inside the `request.Request.Context()` object and can be retrieved using the `context.Client` method.

```
SecureFilter(req *restful.Request, res *restful.Response, chain *restful.FilterChain)
```

#### DynamicFilterGenerator

a `restful.FilterFunction` generator, provided a `schema.GroupVersionKind` is given. Will inject the client in the request's Context and can be retrieved using `context.DynamicClient` method.

```
DynamicFilterGenerator(gvk *schema.GroupVersionKind) restful.FilterFunction
```

## Query

Declares and initiates a `*dataselect.Query` object into the `request.Request.Context()`

**Constructor:**: `NewQuery()`

### Methods

#### Build
 
Used to injects a Filter inside a *restful.RouteBuild and add the query parameters and its descriptions. To be used when declaring a new `Route`

```
Build(builder *restful.RouteBuilder) *restful.RouteBuilder
```

#### Filter

a `restful.FilterFunction` used in the `Build` method to create and inject the `*dataselect.Query` object

```
Filter(req *restful.Request, res *restful.Response, chain *restful.FilterChain)
```

#### QueryItems

A simplified method to provide a `slice` of objects that implements `metav1.Object` interface and execute filtering, sorting, and pagging. Returns a `[]metav1.Object` and `int` 

```
QueryItems(items interface{}, query *dataselect.Query) (result []metav1.Object, count int)
```

#### AddItemCountHeader

Adds the `count` result from `QueryItems` function to the `restful.Response` header.

```
AddItemCountHeader(res *restful.Response, count int)
```

## Other files

### `webservice.go`

Contains multiple helper methods to create a new `*restful.WebService` and to build `Routes`'s documentation

### `rewrite.go`

Custom implementation of the `restful.Router` to allow a path prefix to be added for all `*restful.WebService`s