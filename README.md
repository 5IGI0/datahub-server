# datahub-server

central server for managing/aggregating all data received from the various sources that make up the datahub.

## TODOs

- [X] front
- [ ] token + permissions
- [ ] websocket for dispatching job to workers
- [X] use ORM or QueryBuilder
- [ ] stop parsing json server-side (but rather use "GENERATED ALWAYS AS" column)

unit test
- [ ] check that client-side's hash_ids match database ones

http servers

- [X] store headers
- [X] search by header
- [X] store html meta
- [X] search by html meta
- [X] store robots.txt directives
- [X] search by robots.txt directive
- [X] store certificates
- [X] search by certificate

http services to scrape

- [ ] wordpress
- [ ] phpbb
- [ ] gitea/gogs/forgejo

links to scrape

- [ ] discord invite
- [ ] telegram link
- [ ] linktree
- [X] discourse-based forums
- [ ] ...

...