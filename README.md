# Jackal Storage Outpost

The outpost wraps the controller functionality from the interchain accounts standard. This contract can port all of canine-chain's functionality over to any cosmos chain, so long as that chain supports cosmwasm.

A topology diagram of a file's trip starting with our javascript SDK, through the outpost contract, before arriving at canine-chain is below.

[Storage Outpost Topology Diagram](https://www.figma.com/community/file/1335344086198579221/storage-outpost-right-to-left)

A filetree msg has been successfully sent from wasmd to canined in the e2e environment.

A Complete road map to mainnet and Dapp integration documentation incoming soon.

## Contributors

The outpost contract is an implementation of the cw-ica-controller contract, found here:

[cw-ica-controller GitHub](https://github.com/srdtrk/cw-ica-controller)

The Jackal team thanks Serdar Turkmenafsar for developing the cw-ica-controller, and for his continued support in integrating it with canine-chain modules and jackal.js.

Special shout out to @Reecepbcups for his support during our initial testing of the cw-ica-controller on Juno’s testnet.
