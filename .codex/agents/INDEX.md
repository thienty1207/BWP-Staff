<!-- BARON:ROUTING:START -->
# Baron Agent Routing

Use the three core quality agents as gates, not as workflow owners. Do not dispatch agents recursively.

Run `baron control-plane route "<task>"` before dispatch. After a gate actually runs, record evidence with `baron control-plane record-gate`.

| Agent | Ownership | Trigger | Exclusion | Evidence | Conflicts |
| --- | --- | --- | --- | --- | --- |
| `code-reviewer` | core quality gate | meaningful code change, medium/high-risk work | pure docs/status-only updates unless requested | findings or no-issue review with files/proof/trace gaps | must not plan, implement, or call subagents |
| `security-auditor` | core security gate | auth, permission, tenant/RLS, secrets, upload, payment, dependency, security-sensitive work | non-security low-risk work | severity, evidence, impact, fix, verification | must not provide weaponized exploit steps or call subagents |
| `test-engineer` | core verification gate | implementation, bugfix, release, proof, regression concern | none for meaningful implementation | exact commands, outcomes, missing coverage | must not replace actual test/proof execution |
| `web-performance-auditor` | optional web performance gate | Core Web Vitals, Lighthouse, LCP, INP, CLS, bundle/loading/rendering performance | non-web or non-performance tasks | metric source or potential-impact label | optional web performance only; not included in mandatory gates |
<!-- BARON:ROUTING:END -->

## Custom Agents

Register optional project-specific agents below without replacing the core gates.
