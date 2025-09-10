package common

import (
	"encoding/json"
	"fmt"
)

type Message struct {
    State  int
    Action int
    Data   json.RawMessage
}

func SendRequest(encoder *json.Encoder, state int, action int, payload ...int) int {
	var data []byte = nil;
	var err error
	if payload!=nil {
		data, err = json.Marshal(payload)
		if err != nil {
			fmt.Println("Erro ao serializar payload:", err)
			return 1
		}	
	}
	

	req := Message{
		Action: action,
        State:  state,
		Data:   data,
	}

	err = encoder.Encode(req)
	if err != nil {
		return 2
		//fmt.Println("Erro ao enviar requisição:", err)
		//return 2
	}
	
	return 0
}

func SendRequestList(encoder *json.Encoder, state int, action int, newpayload []int) int {
	var data []byte = nil;
	var err error
	
	data, err = json.Marshal(newpayload)
	if err != nil {
		fmt.Println("Erro ao serializar payload:", err)
		return 1
	}	

	

	req := Message{
		Action: action,
        State:  state,
		Data:   data,
	}

	err = encoder.Encode(req)
	if err != nil {
		return 2
		//fmt.Println("Erro ao enviar requisição:", err)
		//return 2
	}
	
	return 0
}

func ReadData(dec *json.Decoder, msg *Message) ([]any, int) {
	var temp []any
	err := dec.Decode(msg)
	if err!=nil {
		return temp, 1
	}
	
	err = json.Unmarshal(msg.Data, &temp)
	if err!=nil {
		return temp, 2
	}
	return temp, 0
}

func ToInt(v any) int {
    switch x := v.(type) {
    case float64:
        return int(x)
    case int:
        return x
    default:
        return 0
    }
}