"""Regenerate Go data structures from the supplied Java DTOs and entities."""
from pathlib import Path
import re
root=Path(__file__).resolve().parents[2]
base=root/'java/src/main/java'
paths=[]
for folder in ['in/codifi/basket/entity/primary','in/codifi/basket/entity/logs','in/codifi/basket/model','in/codifi/basket/ws/model','in/codifi/cache/model','com/sas/dto']:
 paths+=list((base/folder).rglob('*.java'))
names={p.stem for p in paths}
def typ(t):
 t=t.strip()
 if t.endswith('[]'):return '[]'+typ(t[:-2]).lstrip('*')
 if t.startswith('List<'):return '[]'+typ(t[5:-1]).lstrip('*')
 return {'String':'*string','Date':'*Timestamp','Timestamp':'*Timestamp','LocalDateTime':'*Timestamp','long':'int64','Long':'*int64','int':'int','Integer':'*int','double':'float64','Double':'*float64','float':'float32','Float':'*float32','boolean':'bool','Boolean':'*bool','BigDecimal':'*string','PublicKey':'any','PrivateKey':'any','Object':'any','JSONObject':'map[string]any','JSONArray':'[]any'}.get(t,t if t in names else 'any')
lines=['// Generated from Java DTO/entity declarations by tools/generate_models.py; review business logic separately.','package model','']
entities=[]
for p in sorted(paths):
 s=re.sub(r'/\*.*?\*/','',p.read_text(),flags=re.S);s=re.sub(r'//[^\n]*','',s)
 cls=p.stem;lines+=['// '+cls+' corresponds to '+str(p.relative_to(base))+'.','type '+cls+' struct {']
 if re.search(r'extends CommonEntity',s):lines+=[' CommonEntity']
 start=s.find('{',s.find('class '));s=s[start+1:]
 prev=0
 for m in re.finditer(r'(?:private|public|protected)?\s+([\w.]+(?:<[^;\n]+?>)?(?:\[\])?)\s+(\w+)\s*(?:=\s*[^;]+)?;',s):
  t,n=m.group(1,2);t=t.removeprefix("java.time.").removeprefix("java.sql.").removeprefix("java.util.");anno=s[prev:m.start()];prev=m.end()
  if n=='serialVersionUID' or t not in names and t not in ['String','Date','Timestamp','LocalDateTime','long','Long','int','Integer','double','Double','float','Float','boolean','Boolean','BigDecimal','Object','JSONObject','JSONArray','PublicKey','PrivateKey'] and not t.startswith('List<') and not t.endswith('[]'):continue
  field=n[0].upper()+n[1:];key=n[0].lower()+n[1:];tags=[]
  col=re.search(r'@Column\s*\(\s*name\s*=\s*"([^"]+)"',anno)
  if col:tags+=['column:'+col[1].lower()]
  if '@Id' in anno:tags+=['primaryKey','autoIncrement']
  if '@CreationTimestamp' in anno:tags+=['autoCreateTime']
  if '@UpdateTimestamp' in anno:tags+=['autoUpdateTime']
  if n=='activeStatus' and re.search(r'=\s*1\s*;',m[0]):tags+=['default:1']
  if n=='isExecuted':tags+=['default:0']
  if 'WRITE_ONLY' in anno or '@JsonIgnore' in anno:key='-'
  jp=re.search(r'@JsonProperty\("([^"]+)"\)',anno)
  if jp:key=jp[1]
  if n=='basketScrip':tags=['foreignKey:BasketId;references:BasketId;constraint:OnDelete:CASCADE']
  if 'entity' in str(p) and t=='String' and any(x in n.lower() for x in ['response','request','description','message','attachment','recommendation','scripdetails']):tags+=['type:text']
  tag='json:"'+key+'"'
  if tags:tag+=' gorm:"'+';'.join(tags)+'"'
  lines+=[' '+field+' '+typ(t)+' `'+tag+'`']
 lines+=['}','']
 raw=p.read_text();table=re.search(r'@(?:Table|Entity)\(name\s*=\s*"([^"]+)"',raw)
 if table:
  entities.append(cls);lines+=['func ('+cls+') TableName() string { return "'+table[1].lower()+'" }','']
lines+=['func Entities() []any { return []any{'+','.join('&'+n+'{}' for n in entities)+'} }']
(root/'go/internal/model/java_models.go').write_text('\n'.join(lines)+'\n')
print(len(paths),'models;',len(entities),'entities')
