# Generating TS interfaces

Ensure you have ts-codegen installed:

```
npm install @cosmwasm/ts-codegen
```

Update the interfaces by running the following commands from root:

```
cargo schema
```

```
cosmwasm-ts-codegen generate \
          --plugin client \
          --plugin recoil \
          --schema ./schema \
          --out ./ts \
          --name storage-outpost
```


