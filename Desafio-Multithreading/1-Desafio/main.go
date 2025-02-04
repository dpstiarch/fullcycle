package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type CepViaBrasilResponse struct {
	Cep        string `json:"cep"`
	Logradouro string `json:"logradouro"`
	Bairro     string `json:"bairro"`
	Cidade     string `json:"localidade"`
	Estado     string `json:"uf"`
}

type CepBrasilApiResponse struct {
	Cep        string `json:"cep"`
	Logradouro string `json:"street"`
	Bairro     string `json:"neighborhood"`
	Cidade     string `json:"city"`
	Estado     string `json:"state"`
}

type CepGenericResponse struct {
	Origem    string
	ViaCEP    *CepViaBrasilResponse
	BrasilAPI *CepBrasilApiResponse
}

func taskBrasilApi(cep string, wg *sync.WaitGroup, ch chan<- CepGenericResponse) {

	defer wg.Done()

	url := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)
	resp, err := http.Get(url)

	if err != nil {
		fmt.Println("Erro ao fazer requisição:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var retorno CepBrasilApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&retorno); err != nil {
		fmt.Println("Erro ao decodificar JSON:", err)
		os.Exit(1)
	}

	ch <- CepGenericResponse{
		Origem:    "BrasilAPI",
		BrasilAPI: &retorno,
	}

}

func taskViaCep(cep string, wg *sync.WaitGroup, ch chan<- CepGenericResponse) {

	defer wg.Done()

	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", cep)
	resp, err := http.Get(url)

	if err != nil {
		fmt.Println("Erro ao fazer requisição:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var retorno CepViaBrasilResponse
	if err := json.NewDecoder(resp.Body).Decode(&retorno); err != nil {
		fmt.Println("Erro ao decodificar JSON:", err)
		os.Exit(1)
	}

	ch <- CepGenericResponse{
		Origem: "ViaBrasil",
		ViaCEP: &retorno,
	}

}

func main() {
	var cep string = "71503507"
	var wg sync.WaitGroup
	tempoInicial := time.Now()

	ch := make(chan CepGenericResponse, 2)
	wg.Add(2)
	go taskViaCep(cep, &wg, ch)
	go taskBrasilApi(cep, &wg, ch)

	retorno := <-ch

	if retorno.Origem == "BrasilAPI" {

		fmt.Printf("["+retorno.Origem+"] Finalizou primeiro em %v\n", time.Since(tempoInicial))
		fmt.Printf("CEP: %s\nLogradouro: %s\nBairro: %s\nCidade: %s\nEstado: %s\n",
			retorno.BrasilAPI.Cep, retorno.BrasilAPI.Logradouro, retorno.BrasilAPI.Bairro, retorno.BrasilAPI.Cidade,
			retorno.BrasilAPI.Estado)
	} else {
		fmt.Printf("["+retorno.Origem+"] Finalizou primeiro em %v\n", time.Since(tempoInicial))
		fmt.Printf("CEP: %s\nLogradouro: %s\nBairro: %s\nCidade: %s\nEstado: %s\n",
			retorno.ViaCEP.Cep, retorno.ViaCEP.Logradouro, retorno.ViaCEP.Bairro, retorno.ViaCEP.Cidade, retorno.ViaCEP.Estado)
	}
	fmt.Println("================================================================")
	fmt.Println("")
	wg.Wait() // Aguarda todas as goroutines terminarem antes de encerrar o programa

}
