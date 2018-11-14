package handler

import (
	"github.com/emicklei/go-restful"
	"github.com/wzt3309/k8sconsole/src/app/backend/api"
	"github.com/wzt3309/k8sconsole/src/app/backend/auth"
	authApi "github.com/wzt3309/k8sconsole/src/app/backend/auth/api"
	clientApi "github.com/wzt3309/k8sconsole/src/app/backend/client/api"
	kcErrors "github.com/wzt3309/k8sconsole/src/app/backend/errors"
	"github.com/wzt3309/k8sconsole/src/app/backend/resource/common"
	"github.com/wzt3309/k8sconsole/src/app/backend/resource/dataselect"
	"github.com/wzt3309/k8sconsole/src/app/backend/resource/event"
	"github.com/wzt3309/k8sconsole/src/app/backend/resource/node"
	"github.com/wzt3309/k8sconsole/src/app/backend/resource/pod"
	"golang.org/x/net/xsrftoken"
	"net/http"
	"strconv"
	"strings"
)

type APIHandler struct {
	cManager clientApi.ClientManager
}

// CreateHTTPAPIHandler creates a new HTTP handler that handles all requests to the API of the backend.
func CreateHTTPAPIHandler(cManager clientApi.ClientManager, authManager authApi.AuthManager) (http.Handler, error) {
	apiHandler := APIHandler{cManager: cManager}

	wsContainer := restful.NewContainer()
	wsContainer.EnableContentEncoding(true)

	apiV1Ws := new(restful.WebService)
	apiV1Ws.Path("/api/v1").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	wsContainer.Add(apiV1Ws)

	authHandler := auth.NewAuthHandler(authManager)
	authHandler.Install(apiV1Ws)

	apiV1Ws.Route(
		apiV1Ws.GET("csrftoken/{action}").
			To(apiHandler.handleGetCsrfToken).
			Writes(api.CsrfToken{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/pod").
			To(apiHandler.handleGetPods).
			Writes(pod.PodList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/pod/{namespace}").
			To(apiHandler.handleGetPods).
			Writes(pod.PodList{}))

	apiV1Ws.Route(
		apiV1Ws.GET("/node").
			To(apiHandler.handleGetNodeList).
			Writes(node.NodeList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/node/{name}").
			To(apiHandler.handleGetNodeDetail).
			Writes(node.NodeDetail{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/node/{name}/event").
			To(apiHandler.handleGetNodeEvents).
			Writes(common.EventList{}))
	apiV1Ws.Route(
		apiV1Ws.GET("/node/{name}/pod").
			To(apiHandler.handleGetNodePods).
			Writes(pod.PodList{}))

	return wsContainer, nil
}

func (apiHandler *APIHandler) handleGetCsrfToken(request *restful.Request, response *restful.Response) {
	action := request.PathParameter("action")
	token := xsrftoken.Generate(apiHandler.cManager.CSRFKey(), "none", action)
	response.WriteHeaderAndEntity(http.StatusOK, api.CsrfToken{Token: token})
}

func (apiHandler *APIHandler) handleGetPods(request *restful.Request, response *restful.Response) {
	k8sClient, err := apiHandler.cManager.Client(request)
	if err != nil {
		kcErrors.HandleInternalError(response, err)
		return
	}

	namespace := parseNamespacePathParameter(request)
	dsQuery := parseDataSelectPathParameter(request)
	result, err := pod.GetPodList(k8sClient, namespace, dsQuery)
	if err != nil {
		kcErrors.HandleInternalError(response, err)
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetNodeList(request *restful.Request, response *restful.Response) {
	k8sclient, err := apiHandler.cManager.Client(request)
	if err != nil {
		kcErrors.HandleInternalError(response, err)
		return
	}

	dsQuery := parseDataSelectPathParameter(request)
	result, err := node.GetNodeList(k8sclient, dsQuery)
	if err != nil {
		kcErrors.HandleInternalError(response, err)
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetNodeDetail(request *restful.Request, response *restful.Response) {
	k8sClient, err := apiHandler.cManager.Client(request)
	if err != nil {
		kcErrors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("name")
	dsQuery := parseDataSelectPathParameter(request)
	result, err := node.GetNodeDetail(k8sClient, name, dsQuery)
	if err != nil {
		kcErrors.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetNodeEvents(request *restful.Request, response *restful.Response) {
	k8sClient, err := apiHandler.cManager.Client(request)
	if err != nil {
		kcErrors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("name")
	dsQuery := parseDataSelectPathParameter(request)
	result, err := event.GetNodeEvents(k8sClient, dsQuery, name)
	if err != nil {
		kcErrors.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

func (apiHandler *APIHandler) handleGetNodePods(request *restful.Request, response *restful.Response) {
	k8sClient, err := apiHandler.cManager.Client(request)
	if err != nil {
		kcErrors.HandleInternalError(response, err)
		return
	}

	name := request.PathParameter("name")
	dsQuery := parseDataSelectPathParameter(request)
	result, err := node.GetNodePods(k8sClient, dsQuery, name)
	if err != nil {
		kcErrors.HandleInternalError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, result)
}

// Get namespaces from path parameter
func parseNamespacePathParameter(request *restful.Request) *common.NamespaceQuery {
	namespace := request.PathParameter("namespace")
	namespaces := strings.Split(namespace, ",")
	var noBlankNamespaces []string
	for _, n := range namespaces {
		n = strings.Trim(n, " ")
		if len(n) > 0 {
			noBlankNamespaces = append(noBlankNamespaces, n)
		}
	}

	return common.NewNamespaceQuery(noBlankNamespaces)
}

func parsePaginationPathParameter(request *restful.Request) *dataselect.PaginationQuery {
	itemsPerPage, err := strconv.ParseInt(request.QueryParameter("itemsPerPage"), 10, 0)
	if err != nil {
		return dataselect.NoPagination
	}

	page, err := strconv.ParseInt(request.QueryParameter("page"), 10, 0)
	if err != nil {
		return dataselect.NoPagination
	}

	// Frontend page start from 1 and backend start from 0
	return dataselect.NewPaginationQuery(int(itemsPerPage), int(page - 1))
}

func parseFilterPathParameter(request *restful.Request) *dataselect.FilterQuery {
	return dataselect.NewFilterQuery(strings.Split(request.QueryParameter("filterBy"), ","))
}

func parseSortPathParameter(request *restful.Request) *dataselect.SortQuery {
	return dataselect.NewSortQuery(strings.Split(request.QueryParameter("sortBy"), ","))
}

// Parses query parameters of the request and returns a DataSelectQuery object
func parseDataSelectPathParameter(request *restful.Request) *dataselect.DataSelectQuery {
	paginationQuery := parsePaginationPathParameter(request)
	sortQuery := parseSortPathParameter(request)
	filterQuery := parseFilterPathParameter(request)
	return dataselect.NewDataSelectQuery(paginationQuery, sortQuery, filterQuery)
}