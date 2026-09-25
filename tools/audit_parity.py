"""Compare live Java REST annotations with Go routes and write the review inventory."""
from pathlib import Path
import re,json,sys
root=Path(__file__).resolve().parents[2];base=root/'java/src/main/java';out=root/'go/docs'
def clean(s):
 s=re.sub(r'/\*.*?\*/','',s,flags=re.S)
 return re.sub(r'//[^\n]*','',s)
files={p.stem:(p,clean(p.read_text())) for p in base.rglob('*.java')}
java=[]
for name,(path,s) in files.items():
 if 'public class' not in s:continue
 declaration=re.search(r'public class\s+\w+[^\{]*',s)
 prefix=re.findall(r'@Path\("([^"]*)"\)',s[:declaration.start()])
 if not prefix:continue
 prefix=prefix[-1]
 impl=re.search(r'implements\s+(\w+)',declaration[0]);spec=s
 if impl and impl[1] in files:spec=files[impl[1]][1]
 pattern=r'((?:\s*@\w+(?:\([^;]*?\))?\s*)+)(?:public\s+)?(?:RestResponse<.*?>|JSONObject|String)\s+(\w+)\s*\('
 for m in re.finditer(pattern,spec,re.S):
  annotation=m[1];method=re.search(r'@(GET|POST|DELETE|PUT|PATCH)\b',annotation)
  if not method:continue
  paths=re.findall(r'@Path\("([^"]*)"\)',annotation);suffix=paths[-1] if paths else ''
  endpoint='/'+('/'.join((prefix.strip('/'),suffix.strip('/')))).strip('/')
  java.append({'method':method[1],'path':endpoint,'java_controller':str(path.relative_to(base)),'java_method':m[2]})
go_source=(root/'go/internal/app/app.go').read_text()
go_routes={ (m[1],m[2]):{'module':m[3],'go_action':m[4]} for m in re.finditer(r'\{"(GET|POST|DELETE|PUT|PATCH)", "([^"]+)", "([^"]+)", "([^"]+)"',go_source)}
missing=[]
for r in java:
 k=(r['method'],r['path'])
 if k not in go_routes:missing.append(k)
 else:r.update(go_routes[k])
extra=set(go_routes)-{(r['method'],r['path'])for r in java}
if missing or extra:print('Missing:',missing,'Extra:',sorted(extra));sys.exit(1)
out.mkdir(exist_ok=True);(out/'api-inventory.json').write_text(json.dumps(sorted(java,key=lambda x:x['path']),indent=2)+'\n')
lines=['# API inventory','','Generated from active Java JAX-RS annotations. `python3 tools/audit_parity.py` fails if a Java route lacks a Go route or a Go route lacks a Java counterpart. This checks route presence, not behavioral equivalence.','','| Method | Path | Module | Java method |','|---|---|---|---|']
for r in sorted(java,key=lambda r:r['path']):lines.append(f"| {r['method']} | `{r['path']}` | `{r['module']}` | `{r['java_method']}` |")
(out/'API.md').write_text('\n'.join(lines)+'\n')
print(f'{len(java)} Java routes match {len(go_routes)} Go routes; no missing or extra routes.')
# Every Java source file is accounted for; framework interfaces are consolidated
# into Go handlers/services instead of creating empty one-to-one wrapper files.
modelnames=set(re.findall(r'type (\w+) struct', (root/'go/internal/model/java_models.go').read_text()))
manual={
 'AppConstants':'internal/config/config.go; internal/app/common.go (protocol messages)',
 'PrepareResponse':'internal/app/common.go', 'ValidateUtil':'internal/utility/legacy.go',
 'StringUtil':'internal/utility/legacy.go; standard strings package',
 'CommonUtils':'internal/utility/legacy.go; upstream TLS config',
 'CodifiUtil':'internal/utility/legacy.go; upstream TLS config',
 'AppUtil':'internal/app/auth.go; internal/app/cache.go',
 'DefaultRestController':'internal/app/auth.go', 'TokenResource':'internal/app/app.go (see PARITY.md)',
 'BasketOrderService':'internal/app/basket.go; internal/app/upstream.go',
 'BasketOrderApiService':'internal/app/admin.go', 'IResearchCallApiService':'internal/app/research.go',
 'ThematicBasketService':'internal/app/thematic.go; internal/app/orderbook.go',
 'UserNotificationService':'internal/app/notifications.go',
 'HoldingsController':'internal/app/holdings.go', 'HoldingsDao':'internal/app/holdings.go',
 'ResearchCallDAO':'internal/app/research.go; internal/app/repository.go',
 'ThematicDao':'internal/app/thematic.go; internal/app/repository.go',
 'OrderStatusFeedDAO':'internal/app/orderbook.go',
 'InternalRestService':'internal/app/upstream.go', 'OrdersRestService':'internal/app/orderbook.go',
 'SpanMarginRestService':'internal/app/upstream.go',
 'PushNoficationUtils':'internal/app/notifications.go', 'FcmNotificationUtils':'internal/app/notifications.go',
 'AccessLogFilter':'internal/app/logging.go', 'BaskerOrderFilter':'internal/app/app.go',
 'AccessLogManager':'internal/app/logging.go', 'AccessLogCache':'internal/app/logging.go (synchronous bounded writes)',
 'Scheduler':'internal/app/app.go; internal/app/admin.go', 'InitialLoader':'internal/app/app.go',
 'CacheService':'internal/app/admin.go', 'HazleCacheController':'internal/app/cache.go',
 'Test':'test fixtures only; hardcoded production-device sender intentionally not ported',
}
lines=['# Java source coverage','','This index accounts for source files. It is not a proof of identical behavior; see PARITY.md for limitations and changes.','','| Java source | Go counterpart |','|---|---|']
for p in sorted(base.rglob('*.java')):
 name=p.stem
 if name in modelnames:target='internal/model/java_models.go (and JDBC extensions in tables.go)'
 elif name in manual:target=manual[name]
 elif '/controller/' in str(p) or name=='CacheController':target='internal/app/app.go route table and corresponding module handler'
 elif '/config/' in str(p):target='internal/config/config.go; internal/app/{app,cache,logging,notifications}.go'
 elif '/repository/' in str(p):target='GORM models/queries in corresponding module service; internal/app/repository.go'
 elif '/spec/' in str(p):target='Go module handler/service functions; no framework proxy required'
 else:raise RuntimeError('Unaccounted Java file: '+str(p))
 lines.append(f'| `{p.relative_to(base)}` | {target} |')
(out/'SOURCE_COVERAGE.md').write_text('\n'.join(lines)+'\n')
print(f'{len(list(base.rglob("*.java")))} Java source files accounted for.')
