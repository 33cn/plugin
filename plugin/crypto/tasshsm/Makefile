
proj := "tass"
.PHONY: default build sm2

default: build

build:
	g++ -o tass demo/main.cpp -I include -std=c++11 -Llinux64 -lTassSDF4PCIeSM -lTassAPI4PCIECrypotoCard -ltass_pcie_api

sm2:
	g++ -o test_sm2 sm2/main.cpp -I include -std=c++11 -Llinux64 -lTassSDF4PCIeSM -lTassAPI4PCIECrypotoCard -ltass_pcie_api
