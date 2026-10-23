import {sampleDocument,sampleSchema,Schema} from './logic';
// Development-only browser preview. Production requires the native Wails bridge.
export function previewAPI(){let workspace={schema:sampleSchema(),theme:'dark',document_path:'',schema_path:''};let connected=false;return {
 Bootstrap:async()=>({workspace,document:sampleDocument,notice:'UI preview: fictional sample. No live model requests.',connection:{account:'',configured:connected,source:'preview',message:'Development preview; no Cloudflare requests.'}}),
 Ready:async()=>{},UpdateWorkspace:async(w:any)=>{workspace=w;},Validate:async()=>'',
 ConfigureConnection:async()=>{connected=true;return {account:'a'.repeat(32),configured:true,source:'preview',message:'Synthetic preview connection.'};},
 ForgetConnection:async()=>({account:'',configured:false,source:'preview',message:'Preview disconnected.'}),TestConnection:async()=>'UI preview only; no request made.',
 Extract:async(_s:string,schema:Schema)=>{await new Promise(resolve=>setTimeout(resolve,1100));const answers:any={};const sample:any={invoice_number:'INV-2291',vendor:'Northwind Supply',total:1080,due_date:'2026-10-02',po_number:null};for(const f of schema.fields){const value=sample[f.name]??null;answers[f.name]={type:f.type,value,confidence:.99};if(value==='INV-2291')answers[f.name].span=[8,16];if(value==='Northwind Supply')answers[f.name].span=[22,38];}return {answers,truncated:false,usage:{input_tokens:655}};},
 CancelExtraction:async()=>{},OpenDocument:async()=>({path:'',text:''}),OpenSchema:async()=>({path:'',schema:sampleSchema()}),SaveSchema:async()=>'',
 GenerateCode:async(language:string)=>({language,code:'// Development preview. Run the desktop application to generate real code.',note:'Synthetic browser preview; no backend or model calls.'}),SaveCode:async()=>'',CopyText:async(text:string)=>navigator.clipboard.writeText(text),
};}
