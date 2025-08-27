#!/bin/bash

for i in $(seq 1 50)
do
  gnome-terminal -- bash -c "cd ~/Documentos/MI_CONC_CONECT/P1/teste1/client && go run client.go; exec bash"
done
