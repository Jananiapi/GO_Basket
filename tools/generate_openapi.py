from pathlib import Path
import json,re
from api_reference_data import DATA
root=Path(__file__).resolve().parents[1]
routes=json.loads((root/'docs/api-inventory.json').read_text())
models=(root/'internal/model/java_models.go').read_text()
smoke=json.loads((root/'docs/smoke-results.json').read_text())
examples={(r['method'],r['path']):r for r in smoke['routes']}
assert {r['go_action'] for r in routes} == set(DATA)
names=set(re.findall(r'type (\w+) struct',models))
def schema(t):
 nullable=t.startswith('*');t=t.lstrip('*')
 if t.startswith('[]'):return {'type':'array','items':schema(t[2:])}
 if t in names:s={'$ref':'#/components/schemas/'+t}
 elif t=='Timestamp':return {'oneOf':[{'type':'integer','description':'Epoch milliseconds'},{'type':'string','description':'ISO date or date-time accepted on input'},{'type':'null'}]}
 elif t=='string':s={'type':'string'}
 elif t=='bool':s={'type':'boolean'}
 elif t.startswith('int'):s={'type':'integer'}
 elif t.startswith('float'):s={'type':'number'}
 else:return {}
 if nullable:s={'anyOf':[s,{'type':'null'}]}
 return s
schemas={}
for m in re.finditer(r'type (\w+) struct \{(.*?)\n\}',models,re.S):
 props={}
 for f in re.finditer(r'\w+\s+(\S+)\s+`json:"([^",]+)',m[2]):
  if f[2]!='-':props[f[2]]=schema(f[1])
 schemas[m[1]]={'type':'object','properties':props}
schemas['GenericResponse']={'type':'object','required':['status','message','result'],'properties':{'status':{'type':['string','null']},'message':{'type':['string','null']},'result':{}}}
request_types={'create':'BasketOrderReq','rename':'BasketOrderReq','addScrip':'BasketOrderReq','deleteScrip':'BasketOrderReq','updateScrip':'BasketOrderReq','updateScripList':'BasketOrderListUpdateReq','execute':'ExecuteBasketOrderReq','span':'[]SpanMarginReq','nestSpan':'[]BasketMarginRequest','adminCreate':'AdminBasketOrderReq','adminTemp':'AdminBasketOrderReq','researchBasket':'ResearchCallRequest','research':'ResearchCallRequest','reports':'ResearchCallRequest',**{x:'ExecuteBasketOrderReq' for x in ['invest','investV1','review','rebalance','view']}}
spec={'openapi':'3.1.0','info':{'title':'Basket Order Module','version':'1.0.0','description':'Go port of the supplied Java module. See docs/PARITY.md for source defects, sample conflicts, and integration requirements. Business failures generally use HTTP 200 and status Not ok.'},'servers':[{'url':'http://localhost:9009'}],'security':[{'bearerAuth':[]}],'paths':{},'components':{'securitySchemes':{'bearerAuth':{'type':'http','scheme':'bearer','bearerFormat':'JWT'},'adminToken':{'type':'apiKey','in':'header','name':'Authorization'}},'schemas':schemas}}
for r in routes:
 a=r['go_action'];response={'$ref':'#/components/schemas/GenericResponse'}
 if a in ['execute','invest','investV1']:response={'type':'array','items':response}
 op={'operationId':r['java_method'],'tags':[r['module']],'description':'Enabled by modules.'+r['module']+'.','responses':{'200':{'description':'Response (check status for business failures)','content':{'application/json':{'schema':response}}},'401':{'description':'Unauthorized'},'404':{'description':'Route disabled or absent'}}}
 if a in request_types:op['requestBody']={'required':True,'content':{'application/json':{'schema':schema(request_types[a])}}}
 params=re.findall(r'\{(\w+)\}',r['path'])
 if params:op['parameters']=[{'name':p,'in':'path','required':True,'schema':{'type':'integer'}} for p in params]
 if a in ['adminCreate','adminTemp']:op['security']=[{'adminToken':[]}]
 if a=='deleteExpired':op['security']=[]
 if a=='category':op['responses']={'204':{'description':'The Java service returns null; no category-header implementation was supplied.'}}
 if a in ['token','logout']:op['responses']['200']['content']={'text/html' if a=='token' else 'text/plain':{'schema':{'type':'string'}}}
 d=DATA[a]
 op['summary']=d['title']
 op['description'] += ' '+d['notes']
 if d['fields']:
  op['description'] += '\n\nRequest fields:\n'+'\n'.join(f'- {f} ({typ}; {required}): {desc}' for f,typ,required,desc in d['fields'])
 if d['errors']:op['description'] += '\n\nBusiness messages: '+', '.join(d['errors'])
 op['externalDocs']={'description':'Complete API reference','url':'API_REFERENCE.md#op-'+a.lower()}
 if d['body'] is not None:
  op['requestBody']['content']['application/json']['example']=d['body']
  # Narrow each request to fields used by this operation; shared DTOs contain
  # unused legacy fields and do not by themselves express validation rules.
  original=op['requestBody']['content']['application/json']['schema']
  array=original.get('type')=='array'
  ref=original['items'] if array else original
  props=schemas[ref['$ref'].split('/')[-1]]['properties']
  fields={f:dict(props.get(f,{}),description=desc+' Requirement: '+required+'.') for f,typ,required,desc in d['fields']}
  required=[f for f,typ,req,desc in d['fields'] if req in ['Required','Required per item'] or (req=='Required in V1' and a=='investV1')]
  body_schema={'type':'object','properties':fields}
  if required:body_schema['required']=required
  if array:body_schema={'type':'array','items':body_schema}
  op['requestBody']['content']['application/json']['schema']=body_schema
 example=examples[(r['method'],r['path'])]
 if a!='category':
  media=next(iter(op['responses']['200']['content'].values()))
  raw=example['response']
  media['example']=raw.strip() if a in ['token','logout'] else json.loads(raw)
 spec['paths'].setdefault(r['path'],{})[r['method'].lower()]=op
(root/'docs/openapi.json').write_text(json.dumps(spec,indent=2)+'\n')
print('OpenAPI contains',len(routes),'operations and',len(schemas),'schemas.')
