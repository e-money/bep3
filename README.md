# BEP3 for Cosmos SDK

This is a fork of the [Kava](https://www.kava.io/) project's [BEP3](https://github.com/binance-chain/BEPs/blob/master/BEP3.md) module implementation.

The fork is made from release [v0.12.0](https://github.com/Kava-Labs/kava/tree/v0.12.0). Its main purpose is to decouple the module from the larger Kava application and make it easier to use in other Cosmos SDK based blockchains.

Notable changes to this fork:

- Upgraded to the breaking Cosmos SDK Stargate 0.4x Release i.e. Protobuf serialization, gRPC enhancements.

- Replaced the height lock mechanism with an equivalent in time span (minute as the lowest time unit) for compatibility with chains featuring asynchronous block appends i.e. [Avalanche](https://github.com/ava-labs/avalanchego/)
