package decorator

import (
	"fmt"
	"strconv"
	"strings"

	restful "github.com/emicklei/go-restful/v3"
	"gomod.alauda.cn/alauda-backend/pkg/context"
	"gomod.alauda.cn/alauda-backend/pkg/dataselect"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// SortParamKey sort parameter key on query string
	SortParamKey = "sortBy"
	// FilterParamKey filter parameter key on query string
	FilterParamKey = "filterBy"
	// PageSizeKey page size parameter key on query string
	PageSizeKey = "itemsPerPage"
	// PageNumberKey current page number parameter key on query string
	PageNumberKey = "page"
	// HeadersItemsCount header item for items count in response
	HeadersItemsCount = "items_count"
)

// Query query api decorator
// will inject a *dataselect.Query on context for posterior use
type Query struct {
	SortParamKey      string
	FilterParamKey    string
	PageSizeKey       string
	PageNumberKey     string
	HeadersItemsCount string
}

// NewQuery query decorator for api
func NewQuery() Query {
	return Query{
		SortParamKey:      SortParamKey,
		FilterParamKey:    FilterParamKey,
		PageSizeKey:       PageSizeKey,
		PageNumberKey:     PageNumberKey,
		HeadersItemsCount: HeadersItemsCount,
	}
}

// Build adds all parameters to the *restful.RouteBuilder and injects Query.Filter as
// a Filter in in
func (q Query) Build(builder *restful.RouteBuilder) *restful.RouteBuilder {
	return builder.Filter(q.Filter).
		Param(
			restful.QueryParameter(q.SortParamKey, `Sorting parameter for query in the format: "a,property_name" where a stands for ascending or descending (only "a" or "d" can be used), property_name stands for the property to be used, i.e name, metadata, creationTimestamp, labels, annotations`).
				AllowMultiple(true),
		).
		Param(
			restful.QueryParameter(q.FilterParamKey, `Filter parameter for query in the format: "property_name,value" where property_name stands for the property to be used, i.e name, metadata, creationTimestamp, labels, annotations and value for the desired value. Different property names uses different filtering methods`).
				AllowMultiple(true),
		).
		Param(
			restful.QueryParameter(q.PageSizeKey, "Page size for pagging in query. Must be used together with "+q.PageNumberKey).
				AllowMultiple(false).
				DataFormat("integer"),
		).
		Param(
			restful.QueryParameter(q.PageNumberKey, "Page for pagging in query. Must be used together with "+q.PageSizeKey).
				AllowMultiple(false).
				DataFormat("integer"),
		)
}

// Filter middleware to handle Query requests
func (q Query) Filter(req *restful.Request, res *restful.Response, chain *restful.FilterChain) {
	query := dataselect.NewDataSelectQuery(dataselect.NoPagination, dataselect.NoSort, dataselect.NoFilter)

	var (
		value                string
		page, pageSize       int64
		pageErr, pageSizeErr error
	)
	// filter
	if value = req.QueryParameter(q.FilterParamKey); value != "" {
		query.FilterQuery = dataselect.NewFilterQuery(strings.Split(value, ","))
	}
	// sort
	if value = req.QueryParameter(q.SortParamKey); value != "" {
		query.SortQuery = dataselect.NewSortQuery(strings.Split(value, ","))
	}

	// pagination
	pageSize, pageErr = strconv.ParseInt(req.QueryParameter(q.PageSizeKey), 10, 0)
	page, pageSizeErr = strconv.ParseInt(req.QueryParameter(q.PageNumberKey), 10, 0)
	if pageSizeErr == nil && pageErr == nil {
		query.PaginationQuery = dataselect.NewPaginationQuery(int(pageSize), int(page-1))
	}
	if query == nil {
		// preventing the context to be empty with a standard NoFilter
		query = dataselect.NoDataSelect
	}
	req.Request = req.Request.WithContext(context.WithQuery(req.Request.Context(), query))
	chain.ProcessFilter(req, res)
}

// QueryItems returns a query for items using standard ObjectDataCell (for metav1.Object interface)
func (q Query) QueryItems(items interface{}, query *dataselect.Query) (result []metav1.Object, count int) {
	if items != nil && query != nil {
		itemCells := dataselect.ToObjectCellSlice(items)
		itemCells, count = dataselect.GenericDataSelectWithFilter(itemCells, query)
		result = dataselect.FromCellToObjectSlice(itemCells)
	} else {
		result = []metav1.Object{}
	}
	return
}

// AddItemCountHeader adds a count to the response Header
func (q Query) AddItemCountHeader(res *restful.Response, count int) {
	res.Header().Add(q.HeadersItemsCount, fmt.Sprintf("%d", count))
}
