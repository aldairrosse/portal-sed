# Tasks: Add Pillar Type

- [ ] 1. Ent schema: add `type` field + `go generate`
- [ ] 2. Migration `000020_add_pillar_type`
- [ ] 3. DTOs: add `Type` to PillarListItem, PillarDetail, Create/UpdateRequest
- [ ] 4. Service + repo: filter by type, set default, validation on type change
- [ ] 5. Handler: parse `?type=` query param
- [ ] 6. OpenAPI spec update
- [ ] 7. Frontend: types, store, page (refetch on tab switch)
- [ ] 8. Verify: `go test` + `pnpm build`
