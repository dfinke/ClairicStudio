export type Field = {name:string;type:string;instructions:string;settings:Record<string,any>};
export type Schema = {version:number;fields:Field[];reference_date?:string;include_quantiles?:boolean};
export type Result = {answers:Record<string,any>;truncated:boolean;state_text?:string;usage?:Record<string,any>};
export const types=['text','number','date','time','datetime','duration','list','estimate','forecast'];
export const sampleDocument='Invoice INV-2291 from Northwind Supply, issued 2026-09-02, net 30. Four pallets at $250.00 each, plus $80.00 shipping. Contact: Dana Ruiz, Priya Shah.';
export function newField(name:string,type='text'):Field{
 const settings:Record<string,any>={};if(type==='list')settings.items='text';
 if(type==='estimate'||type==='forecast')Object.assign(settings,{min:'0',max:'100',step:'1'});
 if(type==='forecast')Object.assign(settings,{horizon:'4',interval:'week'});
 return {name,type,instructions:'',settings};
}
export function sampleSchema():Schema{return {version:1,fields:[
 {...newField('invoice_number'),instructions:'Complete invoice number'},
 {...newField('vendor'),instructions:'Full supplier name'},
 {...newField('total','number'),instructions:'Total amount due including shipping',settings:{unit:'USD',decimals:'2'}},
 {...newField('due_date','date'),instructions:'Issue date plus the net payment days'},
 {...newField('po_number'),instructions:'Purchase order number; null if absent'},
]};}
export function settingsFor(field:Field):string[]{
 if(['text','date','time','datetime'].includes(field.type))return ['null_threshold'];
 if(['number','duration'].includes(field.type))return ['unit','integer','decimals','null_threshold'];
 if(field.type==='list')return field.settings.items==='number'?['items','unit','decimals']:['items'];
 return ['min','max','step','scale','unit',...(field.type==='forecast'?['horizon','interval']:[])];
}
export function localError(schema:Schema,document:string,requireDocument=true):string{
 if(requireDocument&&!document.trim())return 'Paste text or open a document to preview.';
 if(!schema.fields.length)return 'Add a field to define what to extract.';
 const seen=new Set<string>();
 for(const f of schema.fields){
  if(!/^[A-Za-z_][A-Za-z0-9_.\-]{0,63}$/.test(f.name))return 'Field names need 1–64 letters, digits, underscores, dots or hyphens; start with a letter or underscore.';
  if(seen.has(f.name.toLowerCase()))return 'Each field needs a unique name.';seen.add(f.name.toLowerCase());
  if(!types.includes(f.type))return 'Choose a supported field type.';
  if(f.type==='list'&&!['text','number','date'].includes(f.settings.items))return `${f.name}: choose a list item type.`;
  if(['estimate','forecast'].includes(f.type)&&(!String(f.settings.min??'').trim()||!String(f.settings.max??'').trim()))return `${f.name}: minimum and maximum are required.`;
 }
 return '';
}
// Providers may use code-point or UTF-16 offsets. Highlight only an exact match.
export function verifiedSpan(document:string,answer:any):[number,number]|null{
 if(!answer||typeof answer.value!=='string'||!Array.isArray(answer.span)||answer.span.length!==2)return null;
 const [start,end]=answer.span;if(!Number.isInteger(start)||!Number.isInteger(end)||start<0||end<=start)return null;
 const points=Array.from(document);
 if(end<=points.length&&points.slice(start,end).join('')===answer.value)return [points.slice(0,start).join('').length,points.slice(0,end).join('').length];
 if(end<=document.length&&document.slice(start,end)===answer.value)return [start,end];
 return null;
}
export function valuesOnly(result:Result,fields:Field[]):Record<string,any>{
 const values:Record<string,any>=Object.create(null);for(const f of fields){if(!Object.prototype.hasOwnProperty.call(result.answers,f.name))throw new Error(`Missing answer ${f.name}`);values[f.name]=result.answers[f.name].value;}return values;
}
export function displayValue(value:any):string{return value===null?'Not found':typeof value==='object'?JSON.stringify(value):String(value??'Missing answer');}
