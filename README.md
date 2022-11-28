# Hello OpenBanking stablecoin
An ERC-20 stablecoin system demonstrating on-chain provable mint using OpenBanking.

## Why?

The current crop of fully backed stablecoins raises questions about 2 aspects of their function:  
* opaqueness of the minting process (usually a private OTC transaction), 
* opaqueness of their reserves (putting trust on 3rd party financial audits)

This project is an attempt to address the first one by utilising the OpenBanking flow.  
Recording the interaction on-chain (consent creation and approval), but keeping the contents of it private (bank account 
numbers and names).

## What?

This is a monorepo composed of 2 folders  
* `chain`: the enhanced ERC-20 smart contract, its unit tests and Hardhat tasks
* `tpp-client`: the Golang Ethereum process to automate the miniting/redemption process and OpenBanking flow

The code is built progressively in the different branches (`v1`, `v2`,...), with new functionality layered on.

## How?

The code building stages, design choices and trade-offs are discussed at length in my [technical blog post](https://sgerogia.github.io/Stablecoin-Openbanking/).  
Use the blog post as your code companion.