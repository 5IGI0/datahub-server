# datahub-server

central server for managing/aggregating all data received from the various sources that make up the datahub.

## TODOs

- [ ] front
- [ ] token + permissions
- [ ] websocket for dispatching job to workers
- [X] use ORM or QueryBuilder

unit test
- [ ] check that client-side's hash_ids matches database ones

http servers

- [X] store headers
- [X] search by header
- [X] store html meta
- [X] search by html meta
- [X] store robots.txt directives
- [X] search by robots.txt directive
- [ ] store certificates
- [ ] search by certificate

http services to scrape

- [ ] wordpress
- [ ] phpbb
- [ ] gitea/gogs/forgejo

links to scrape
- [ ] discord invite
- [ ] telegram link
- [ ] linktree
- [ ] ...

...