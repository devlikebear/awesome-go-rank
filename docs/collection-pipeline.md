# Collection pipeline decisions

## GraphQL batch evaluation

Decision: keep REST as the production path for now. Revisit GraphQL behind a
provider boundary after the snapshot history provides a benchmark baseline.

GitHub GraphQL exposes `repository(owner:, name:)` as a singular root field.
Fetching arbitrary repositories in one request therefore requires generating a
query with aliases and mapping partial field errors back to individual
repositories. The documented `1`–`100` constraint applies to connection
pagination, not to a supported "100 arbitrary repositories" batch endpoint.
GraphQL also uses a point-based primary limit and query/node/timeout limits that
are separate from REST limits.

For this one-maintainer project, introducing dynamic query generation and
partial-error recovery at the same time as correctness fixes creates more risk
than the theoretical reduction from roughly 2,700 REST calls to roughly 30
HTTP requests. The bounded REST worker pool remains within the authenticated
REST budget, preserves per-repository retries, and supplies straightforward
failure-rate accounting.

Before switching, run a GraphQL pilot with these acceptance gates:

1. Query 50 aliased repositories per request and map every alias to a full name.
2. Retry missing or errored aliases through the existing REST path.
3. Record GraphQL points, request count, wall time, fallback count, and final
   failure rate for at least seven daily runs.
4. Adopt GraphQL only if it lowers wall time and request pressure without
   increasing missing repositories or hiding partial errors.

References:

- [GitHub GraphQL repository query](https://docs.github.com/en/graphql/reference/repos#repository)
- [GitHub GraphQL rate and query limits](https://docs.github.com/en/graphql/overview/rate-limits-and-query-limits-for-the-graphql-api)

## `pkg/cache` status and use plan

`pkg/cache` is currently not connected to the ranking runtime. It must not be
used to skip daily metadata refreshes because that would turn a current ranking
into a stale one without making the fallback visible.

If the GraphQL pilot is implemented, extend the cache record with `archived`
and use it only as a last-successful-value fallback when both a GraphQL alias
and its REST retry fail. Cache fallbacks must be younger than 24 hours, counted
in the collection summary, and included in failure-threshold reporting. A run
must still fail loudly when fresh plus cached results do not satisfy the
configured threshold. Until those conditions exist, the package remains
isolated and does not influence output.
