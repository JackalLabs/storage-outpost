# Outpost User

A demonstration of how any contract can call the outpost's API in the same transaction. 

This contract can save the address of the outpost to state. 

We can then call the outpost using the exact same API that the outpost itself uses. 

Under the hood, this contract simply makes a cross contract call to its deployed outpost, which
allows for seamless integration with jackal.js, and a smooth UX--i.e., any contract can
perform its core actions, and store data in a single wallet transaction. 

A well annotated e2e test can be found at e2e/interchaintest/

The testing command is:

```

go test -v . -run TestWithContractTestSuite -testify.m TestOutpostUser -timeout 12h

```




