package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tayron/integra-sistema/models"
)

// ProcessarIntegracao -
func ProcessarIntegracao(w http.ResponseWriter, r *http.Request) {
	parametrosURL := mux.Vars(r)
	endpoint := parametrosURL["endpoint"]

	integracao := models.Integracao{}
	integracao = integracao.BuscarPorEndpoint(endpoint)

	if integracao.ID == 0 {
		retornarMensagem("Integracao não encontrada", false, w, r)
	} else {
		processarIntegracao(integracao, w, r)
	}
}

func processarIntegracao(integracao models.Integracao, w http.ResponseWriter, r *http.Request) {
	parametroModel := models.Parametro{}
	id := int64(integracao.ID)
	listaParametro := parametroModel.BuscarPorIDIntegracao(id)

	if integracao.MetodoSistemaOrigem == r.Method {
		switch integracao.MetodoSistemaDestino {
		case "POST":
			mensagem, sucesso := enviarRequisicaoViaPOST(integracao, listaParametro, r)
			retornarMensagem(mensagem, sucesso, w, r)
		case "GET":
			mensagem, sucesso := enviarRequisicaoViaGET(integracao, listaParametro, r)
			retornarMensagem(mensagem, sucesso, w, r)
		}
	} else {
		mensagem := fmt.Sprintf("Método de envio %s não permitido", r.Method)
		retornarMensagem(mensagem, false, w, r)
	}
}

func enviarRequisicaoViaPOST(integracao models.Integracao, listaParametros []models.Parametro, r *http.Request) (string, bool) {

	jsonData := map[string]string{}
	for _, parametro := range listaParametros {
		jsonData[parametro.NomeParametroSaida] = r.FormValue(parametro.NomeParametroEntrada)
	}

	jsonValue, _ := json.Marshal(jsonData)
	response, err := http.Post(integracao.APISistemaDestino, "application/json", bytes.NewBuffer(jsonValue))

	if err != nil {
		return fmt.Sprintf("%s", err), false
	}

	data, _ := ioutil.ReadAll(response.Body)

	return string(data), true
}

func enviarRequisicaoViaGET(integracao models.Integracao, listaParametros []models.Parametro, r *http.Request) (string, bool) {
	req, err := http.NewRequest("GET", integracao.APISistemaDestino, nil)
	if err != nil {
		return fmt.Sprintf("%s", err), false
	}

	query := req.URL.Query()

	for _, parametro := range listaParametros {
		query.Add(parametro.NomeParametroSaida, r.FormValue(parametro.NomeParametroEntrada))
	}

	req.URL.RawQuery = query.Encode()

	resp, err := http.Get(req.URL.String())
	defer resp.Body.Close()

	if err != nil {
		return fmt.Sprintf("%s", err), false
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return fmt.Sprintf("%s", err), false
	}

	stringRetorno := fmt.Sprintf("%s", string(body))

	return stringRetorno, true
}

func retornarMensagem(mensagem string, status bool, w http.ResponseWriter, r *http.Request) {
	parametros := struct {
		Sucesso  bool
		Mensagem string
	}{
		Sucesso:  status,
		Mensagem: mensagem,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(parametros)
}
