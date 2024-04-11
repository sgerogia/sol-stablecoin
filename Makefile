#! /usr/bin/make

build-contract:
	cd chain && npx hardhat compile

test-contract:
	cd chain && npx hardhat test
