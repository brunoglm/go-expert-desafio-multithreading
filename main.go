package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Output struct {
	From    string      `json:"from"`
	Payload interface{} `json:"payload"`
}

type OutputApiCep struct {
	Code       string `json:"code"`
	State      string `json:"state"`
	City       string `json:"city"`
	District   string `json:"district"`
	Address    string `json:"address"`
	Status     int    `json:"status"`
	Ok         bool   `json:"ok"`
	StatusText string `json:"statusText"`
}

type OutputViaCep struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

func getCepByApiCep(cep string, ch chan<- OutputApiCep) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	url := fmt.Sprintf("https://cdn.apicep.com/file/apicep/%s.json", cep)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("Erro na construção da requisição: %s\n", err.Error())
		panic(err)
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Erro ao efetuar a requisição: %s\n", err.Error())
		panic(err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("Erro ao ler a resposta: %s\n", err.Error())
		panic(err)
	}

	var outputApiCep OutputApiCep
	err = json.Unmarshal(body, &outputApiCep)
	if err != nil {
		log.Printf("Erro ao converter o body da resposta: %s\n", err.Error())
		panic(err)
	}
	ch <- outputApiCep
}

func getCepByViaCep(cep string, ch chan<- OutputViaCep) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	url := fmt.Sprintf("http://viacep.com.br/ws/%s/json", cep)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("Erro na construção da requisição: %s\n", err.Error())
		panic(err)
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Erro ao efetuar a requisição: %s\n", err.Error())
		panic(err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("Erro ao ler a resposta: %s\n", err.Error())
		panic(err)
	}

	var outputViaCep OutputViaCep
	err = json.Unmarshal(body, &outputViaCep)
	if err != nil {
		log.Printf("Erro ao converter o body da resposta: %s\n", err.Error())
		panic(err)
	}
	ch <- outputViaCep
}

func main() {
	cep := "12248-610"
	if len(os.Args) > 1 {
		cep = os.Args[1]
	}

	chApiCep := make(chan OutputApiCep)
	chViaCep := make(chan OutputViaCep)

	go getCepByApiCep(cep, chApiCep)

	go getCepByViaCep(cep, chViaCep)

	select {
	case cepOutput := <-chViaCep:
		output := Output{
			From:    "ViaCep",
			Payload: cepOutput,
		}
		json.NewEncoder(os.Stdout).Encode(output)

	case cepOutput := <-chApiCep:
		output := Output{
			From:    "ApiCep",
			Payload: cepOutput,
		}
		json.NewEncoder(os.Stdout).Encode(output)

	case <-time.After(time.Second):
		json.NewEncoder(os.Stderr).Encode("Ocorreu timeout na operação")
	}
}
