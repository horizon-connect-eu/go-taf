#!/bin/bash
#
# Helper script that copies public keys between TAF and AIV *after* having started both applications
#
# Expected directory structure:
#  ├── aiv
#  └── go-taf
#
#
cp ./res/cert/ecdsa_public_key.pem ../aiv/
cp ./res/cert/attestationCertificate.pem ../aiv/
cp  ../aiv/aiv_public_key.pem ./res/cert/