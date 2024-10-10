# E2E Storage Outpost Test

This is a complete end to end test suite which demonstrates the storage outpost being deployed on wasmd (a demo chain) to send 
canine-chain msg types across the relayer to canined (The Jackal Chain).

TODO: better description 

The testing commands are:

```
go test -v . -run TestWithFactoryTestSuite -testify.m TestFactoryCreateOutpost -timeout 12h

go test -v . -run TestWithFactoryTestSuite -testify.m TestMasterMigration -timeout 12h

go test -v . -run TestWithContractTestSuite -testify.m TestIcaContractExecutionTestWithFiletree -timeout 12h

go test -v . -run TestWithContractTestSuite -testify.m TestIcaContractExecutionTestWithOwnership -timeout 12h

go test -v . -run TestWithContractTestSuite -testify.m TestIcaContractExecutionTestWithBuyStorage -timeout 12h

go test -v . -run TestWithContractTestSuite -testify.m TestIcaContractExecutionTestWithMigration  

go test -v . -run TestWithContractTestSuite -testify.m TestPostFile -timeout 12h

go test -v . -run TestWithContractTestSuite -testify.m TestOutpostUser -timeout 12h

go test -v . -run TestWithFactoryClientTestSuite -testify.m TestOutpostFactoryClient -timeout 12h


```

