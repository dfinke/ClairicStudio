import './style.css';
import {createCopyFeedback} from './clipboard';
import {Field,Schema,Result,types,newField,sampleDocument,sampleSchema,settingsFor,localError,verifiedSpan,valuesOnly,displayValue} from './logic';
import {EventsOn} from '../wailsjs/runtime/runtime';

declare global {interface Window {go?:any;runtime?:any}}
type Connection={account:string;configured:boolean;source:string;message:string};
type Workspace={schema:Schema;theme:string;document_path:string;schema_path:string};
const $=<T extends HTMLElement=HTMLElement>(id:string)=>document.getElementById(id) as T;
const clone=<T>(x:T):T=>JSON.parse(JSON.stringify(x));
let api:any=window.go?.main?.App;
const previewMode=import.meta.env.DEV&&new URLSearchParams(location.search).has('preview');
if(!api&&previewMode)api=(await import('./preview')).previewAPI();

document.querySelector('#app')!.innerHTML=`
<header class="app-header"><div class="brand"><svg aria-hidden="true" viewBox="0 0 40 40"><path d="M8 5h17l7 7v23H8z"/><path d="M24 5v9h8M14 21h12M14 27h7"/><circle cx="29" cy="29" r="7"/></svg><div><h1>Clairic <span>Studio</span></h1><p>From documents to working code.</p></div></div>
<div class="header-actions"><button id="connection" class="connection"><span class="dot"></span><span id="connection-label">Connect Cloudflare</span></button><button id="theme" class="icon-button" aria-label="Toggle light and dark appearance" title="Toggle appearance">☀</button><button id="export" class="primary" title="Preview and save a program">Export code <span aria-hidden="true">↗</span></button></div></header>
<nav class="toolbar" aria-label="Workspace actions"><div><button id="new" title="New schema (Ctrl+N)">New schema</button><button id="open-schema" title="Open schema (Ctrl+O)">Open schema…</button><button id="save-schema">Save schema as…</button><button id="undo" disabled>Undo new schema</button></div><div><span id="workspace-note">Changes save automatically</span><button id="example" class="subtle">Load example</button></div></nav>
<div id="preview-notice" class="preview-notice" hidden>UI preview · synthetic data · no model requests</div>
<main id="workbench"><div id="top-panes">
<section class="panel document-panel" aria-labelledby="document-heading"><div class="panel-heading"><div><span class="eyebrow">01 / SOURCE</span><h2 id="document-heading">Document</h2></div><button id="open-document">Open text file…</button></div><div id="document-path" class="path">Paste text below, or open a UTF-8 document.</div><textarea id="document" spellcheck="false" placeholder="Paste an invoice, customer message, or other document…" aria-label="Source document"></textarea><div class="panel-footer"><span id="document-count">0 characters</span><span id="source-note">Click a text result to inspect its source.</span></div></section>
<div id="column-splitter" class="splitter vertical" role="separator" aria-orientation="vertical" tabindex="0" aria-label="Resize document and fields"></div>
<section class="panel fields-panel" aria-labelledby="fields-heading"><div class="panel-heading"><div><span class="eyebrow">02 / DEFINITION</span><h2 id="fields-heading">Fields <span id="field-count" class="count">0</span></h2></div><button id="add-field" title="Add field (Ctrl+Shift+F)">+ Add field</button></div><div class="field-body"><div class="field-sidebar"><div id="field-list" role="listbox" aria-label="Extraction fields"></div><div class="field-actions"><button id="move-up" title="Move field up" aria-label="Move field up">↑</button><button id="move-down" title="Move field down" aria-label="Move field down">↓</button><button id="duplicate" title="Duplicate selected field">Duplicate</button><button id="delete" class="danger" title="Delete selected field" aria-label="Delete selected field">×</button></div></div><div id="field-editor"><div id="field-empty" class="empty-state"><div class="empty-symbol">{ }</div><h3>Define the shape of your answer</h3><p>Add a field, choose a type, and describe what it should contain.</p></div><div id="editor-controls" hidden><div class="input-row"><label><span>Field <u>n</u>ame</span><input id="field-name" autocomplete="off" spellcheck="false" accesskey="n" title="Focus field name (Alt+N)"></label><label>Type<select id="field-type"></select></label></div><label class="instructions-label">What should this field contain?<textarea id="instructions" rows="3" placeholder="Describe what to extract or compute…"></textarea></label><div id="settings" class="settings-grid"></div></div></div></div><div id="validation" class="validation" role="status"></div></section>
</div><div id="row-splitter" class="splitter horizontal" role="separator" aria-orientation="horizontal" tabindex="0" aria-label="Resize upper panes and results"></div>
<section class="panel results-panel" aria-labelledby="results-heading"><div class="panel-heading"><div><span class="eyebrow">03 / EXTRACTION</span><h2 id="results-heading">Results <span id="results-current" class="small-tag">No extraction yet</span></h2></div><div class="result-actions"><label class="toggle"><input id="auto" type="checkbox" checked><span>Auto-preview</span></label><button id="extract" class="primary" title="Run extraction now, with Auto-preview on or off" disabled>Extract now</button><button id="cancel" hidden>Cancel</button><button id="copy-json" disabled>Copy JSON</button></div></div><div class="table-wrap"><table><thead><tr><th>Field</th><th>Type</th><th>Value</th><th>Confidence</th><th>p10 … p90</th><th>Status</th></tr></thead><tbody id="results-body"></tbody></table><div id="results-empty" class="empty-results"><strong>Your extracted values will appear here.</strong><span>Define fields and connect Cloudflare to run a preview.</span></div></div><details id="answer-details"><summary>Selected answer metadata</summary><pre id="metadata">Select a result to inspect confidence, source spans, and prediction statistics.</pre></details></section></main>
<footer id="status" data-state="ready" role="status" aria-live="polite"><span id="status-badge">READY</span><div class="status-center"><span id="status-message">Paste a document and add a field to get started.</span><div id="activity" hidden><i></i></div></div><span id="usage"></span></footer>
<dialog id="connection-dialog"><form id="connection-form"><div class="dialog-heading"><div><span class="eyebrow">CLOUDFLARE WORKERS AI</span><h2>Connect to Cloudflare</h2></div><button type="button" class="icon-button close-dialog" aria-label="Close connection setup">×</button></div><p class="dialog-intro">Use your own account and Workers AI API token. Schema editing and code export work without a connection.</p><label>Account ID<input id="account" autocomplete="off" spellcheck="false" placeholder="32-character account ID"></label><label>API token<input id="token" type="password" autocomplete="off" placeholder="Paste your API token"></label><label class="check"><input id="remember" type="checkbox">Remember in this computer’s credential store</label><p id="connection-state" class="help"></p><p class="help">Environment variables: <code>CLOUDFLARE_ACCOUNT_ID</code> and <code>CLOUDFLARE_AUTH_TOKEN</code>. Tokens are never included in schemas or exported code.</p><div class="connection-links"><button id="setup-help" type="button" class="subtle">Setup instructions</button><button id="forget" type="button" class="subtle">Forget saved credentials</button></div><div id="setup-instructions" class="help-box" hidden>In Cloudflare, copy your account ID and create an API token with Workers AI Read permission for that account. You can also set the two environment variables before launching Studio. Restart after changing environment variables.</div><p id="connection-error" class="dialog-error" role="alert"></p><div class="dialog-footer"><button id="test-connection" type="button" title="Sends one small live model request">Test extraction</button><span class="help">Test sends a small live request.</span><button type="submit" class="primary">Use credentials</button></div></form></dialog>
<dialog id="export-dialog"><div class="dialog-heading"><div><span class="eyebrow">TAKE IT INTO YOUR WORKFLOW</span><h2>Export a working program</h2></div><button class="icon-button close-dialog" aria-label="Close code export">×</button></div><div class="export-controls"><div class="export-language"><label>Language<select id="language"><option>PowerShell</option><option>Python</option><option>TypeScript</option><option>Go</option><option>Rust</option></select></label><label id="powershell-style-control">Export style<select id="powershell-style"><option value="standalone">Standalone HTTP script</option><option value="module">Use PSAIClairic module</option></select></label></div><fieldset><legend>Document input</legend><label class="check"><input type="radio" name="input-mode" value="file" checked>Read from a file when the program runs</label><label class="check"><input type="radio" name="input-mode" value="embed">Embed the current document</label></fieldset></div><p id="powershell-hint" class="help-box">The PSAIClairic style uses reusable fields and PowerShell pipelines, including Excel workflows. Requires PSAIClairic 0.2.0+.</p><p id="export-note" class="help"></p><textarea id="code" readonly spellcheck="false" aria-label="Generated program"></textarea><p id="export-error" class="dialog-error" role="alert"></p><div class="dialog-footer"><span class="help">Uses your environment credentials. No Studio installation needed to run the exported code.</span><button id="copy-code">Copy code</button><button id="save-code" class="primary">Save program…</button></div></dialog>
`;

let schema:Schema={version:1,fields:[]},selected=-1,documentPath='',schemaPath='',theme='dark';
let connection:Connection={account:'',configured:false,source:'',message:''};
let revision=0,lastRevision=-1,result:Result|null=null,busy=false,queued=false,cancelled=false;
let previewTimer:ReturnType<typeof setTimeout>|undefined,exportRevision=0;
let undoSnapshot:{schema:Schema;path:string;selected:number}|null=null;
let saveChain=Promise.resolve(),booting=true;
const doc=$<HTMLTextAreaElement>('document'),auto=$<HTMLInputElement>('auto');
const copyCodeFeedback=createCopyFeedback($('copy-code'));
const copyJSONFeedback=createCopyFeedback($('copy-json'));
const fieldName=$<HTMLInputElement>('field-name'),fieldType=$<HTMLSelectElement>('field-type'),instructions=$<HTMLTextAreaElement>('instructions');
for(const type of types){const o=document.createElement('option');o.value=type;o.textContent=type;fieldType.append(o);}
function setStatus(state:string,text:string){$('status').dataset.state=state;$('status-badge').textContent=({ready:'READY',stale:'OUT OF DATE',busy:'EXTRACTING',success:'UP TO DATE',error:'ERROR',warning:'CHECK INPUT'} as any)[state]??state;$('status-message').textContent=text;$('activity').hidden=state!=='busy';}
function setTheme(next:string){theme=next==='light'?'light':'dark';document.documentElement.dataset.theme=theme;$('theme').textContent=theme==='dark'?'☀':'☾';}
function currentError(requireDocument=true){return localError(schema,doc.value,requireDocument);}
function saveWorkspace(){if(booting||!api)return;const w:Workspace={schema:clone(schema),theme,document_path:documentPath,schema_path:schemaPath};saveChain=saveChain.then(()=>api.UpdateWorkspace(w)).catch((e:any)=>{$('workspace-note').textContent='Autosave unavailable';setStatus(busy?'busy':'warning',message(e));});}
function message(e:any){return String(e instanceof Error?e.message:e).replace(/^Error:\s*/,'');}
function clearHighlight(){doc.setSelectionRange(doc.selectionEnd,doc.selectionEnd);$('source-note').textContent='Click a text result to inspect its source.';}
function availability(){
 const error=currentError();$('validation').textContent=schema.fields.length?currentError(false):'';
 $<HTMLButtonElement>('extract').disabled=busy||!!error||!connection.configured;
 $('results-empty').querySelector('span')!.textContent=error||(!connection.configured?'Connect Cloudflare to extract this document.':auto.checked?'Choose Extract now, or edit the document or fields for an automatic preview.':'Choose Extract now to process this document.');
 $<HTMLButtonElement>('export').disabled=!!currentError(false);$<HTMLButtonElement>('copy-json').disabled=!result||lastRevision!==revision;
 $<HTMLButtonElement>('save-schema').disabled=!schema.fields.length;
 $<HTMLButtonElement>('undo').disabled=!undoSnapshot;
 $<HTMLButtonElement>('duplicate').disabled=selected<0;$<HTMLButtonElement>('delete').disabled=selected<0;
 $<HTMLButtonElement>('move-up').disabled=selected<=0;$<HTMLButtonElement>('move-down').disabled=selected<0||selected>=schema.fields.length-1;
 $('cancel').hidden=!busy;
 $('results-current').textContent=!result?'No extraction yet':lastRevision===revision?'Current':'Previous results · out of date';
 $('results-current').classList.toggle('stale',!!result&&lastRevision!==revision);
 $('document-count').textContent=`${doc.value.length.toLocaleString()} characters`;
 $('document-path').textContent=documentPath||'Pasted or edited text · held in memory only';
 $('field-count').textContent=String(schema.fields.length);
 $('connection-label').textContent=connection.configured?'Cloudflare connected':'Connect Cloudflare';$('connection').classList.toggle('connected',connection.configured);
}
function schedulePreview(){clearTimeout(previewTimer);if(auto.checked&&connection.configured&&!currentError())previewTimer=setTimeout(()=>void extract(),1000);}
function edited(){copyJSONFeedback.reset();revision++;undoSnapshot=null;clearHighlight();availability();saveWorkspace();setStatus(busy?'busy':result?'stale':'ready',busy?'Extracting the previous revision; your latest edits are queued.':result?'Results are out of date. Latest changes are waiting for extraction.':currentError()||'Ready to extract.');queued=busy;schedulePreview();}
function fieldList(){const list=$('field-list');list.replaceChildren();schema.fields.forEach((f,i)=>{const button=document.createElement('button');button.className='field-tile'+(i===selected?' selected':'');button.setAttribute('role','option');button.setAttribute('aria-selected',String(i===selected));const name=document.createElement('strong');name.textContent=f.name||'Unnamed field';const type=document.createElement('span');type.textContent=f.type;button.append(name,type);button.onclick=()=>{selected=i;renderFields();};list.append(button);});}
const labels:Record<string,string>={unit:'Unit',integer:'Whole number',decimals:'Decimals (0–6)',items:'List item type',min:'Minimum / ISO date',max:'Maximum / ISO date',step:'Step / month / year',scale:'Scale',null_threshold:'Null threshold (0–1]',horizon:'Forecast steps (1–64)',interval:'Forecast interval'};
function renderSettings(f:Field){const root=$('settings');root.replaceChildren();for(const key of settingsFor(f)){
 const label=document.createElement('label');label.textContent=labels[key];let control:HTMLInputElement|HTMLSelectElement;
 const choices=key==='items'?['text','number','date']:key==='scale'?['','linear','log']:key==='interval'?['step','minute','hour','day','week','month','quarter','year']:key==='unit'&&f.type==='duration'?['','seconds','minutes','hours','days','weeks','months','years']:null;
 if(choices){const select=document.createElement('select');for(const v of choices){const o=document.createElement('option');o.value=v;o.textContent=v||'Default';select.append(o);}select.value=String(f.settings[key]??'');control=select;}
 else{const input=document.createElement('input');input.type=key==='integer'?'checkbox':'text';input.spellcheck=false;if(key==='integer')input.checked=!!f.settings[key];else input.value=String(f.settings[key]??'');control=input;}
 control.setAttribute('aria-label',labels[key]);control.addEventListener(control.tagName==='SELECT'||key==='integer'?'change':'input',()=>{f.settings[key]=key==='integer'?(control as HTMLInputElement).checked:control.value;if(key==='items')renderSettings(f);edited();});label.append(control);root.append(label);
}}
function renderFields(){fieldList();const f=schema.fields[selected];$('field-empty').hidden=!!f;$('editor-controls').hidden=!f;if(f){fieldName.value=f.name;fieldType.value=f.type;instructions.value=f.instructions;renderSettings(f);}availability();}
function addField(){if(schema.fields.length>=500){setStatus('warning','This workspace supports up to 500 fields.');return;}let i=1;while(schema.fields.some(f=>f.name.toLowerCase()===`field_${i}`))i++;schema.fields.push(newField(`field_${i}`));selected=schema.fields.length-1;edited();renderFields();fieldName.focus();fieldName.select();}
function newSchema(){if(!schema.fields.length&&!result&&!schemaPath)return;undoSnapshot={schema:clone(schema),path:schemaPath,selected};schema={version:1,fields:[]};schemaPath='';selected=-1;revision++;result=null;lastRevision=-1;queued=false;clearTimeout(previewTimer);clearHighlight();renderFields();renderResults();saveWorkspace();setStatus(busy?'busy':'ready','New schema. Document kept. Undo is available until your next edit.');}
function undoNew(){if(!undoSnapshot)return;const old=undoSnapshot;schema=old.schema;schemaPath=old.path;selected=old.selected;undoSnapshot=null;revision++;result=null;lastRevision=-1;renderFields();renderResults();saveWorkspace();schedulePreview();setStatus('ready','Previous schema restored; results will be extracted again.');}
function renderResults(){copyJSONFeedback.reset();const body=$('results-body');body.replaceChildren();$('results-empty').hidden=!!result;if(!result){$('metadata').textContent='Select a result to inspect confidence, source spans, and prediction statistics.';$('usage').textContent='';return;}
 for(const f of schema.fields){const answer=result.answers[f.name];if(!answer)continue;const row=document.createElement('tr');row.tabIndex=0;row.setAttribute('aria-label',`${f.name}: ${displayValue(answer.value)}`);const cells=[f.name,f.type,displayValue(answer.value),typeof answer.confidence==='number'?`${(100*answer.confidence).toFixed(1)}%`:'—',answer.p10!==undefined&&answer.p90!==undefined?`${answer.p10} … ${answer.p90}`:answer.steps?'Per step':'—',answer.value===null?'Not found':['estimate','forecast'].includes(f.type)?'Prediction':'Returned'];for(const text of cells){const td=document.createElement('td');td.textContent=text;row.append(td);}const select=()=>{body.querySelectorAll('tr').forEach(r=>r.classList.remove('selected'));row.classList.add('selected');$('metadata').textContent=JSON.stringify(answer,null,2);clearHighlight();if(lastRevision===revision){const span=verifiedSpan(doc.value,answer);if(span){doc.focus();doc.setSelectionRange(span[0],span[1]);const line=doc.value.slice(0,span[0]).split('\n').length-1;doc.scrollTop=Math.max(0,line*20-doc.clientHeight/2);$('source-note').textContent=`Source verified for ${f.name}`;}else $('source-note').textContent='This value has no matching direct source span.';}};row.onclick=select;row.onkeydown=e=>{if(e.key==='Enter'){e.preventDefault();select();}};body.append(row);}
 const tokens=result.usage?.input_tokens;$('usage').textContent=tokens!==undefined?`${Number(tokens).toLocaleString()} input tokens`:'';
}
async function extract(){if(busy){queued=true;return;}if(currentError()||!connection.configured)return;
 clearTimeout(previewTimer);const rev=revision,state=doc.value,s=clone(schema);busy=true;queued=false;cancelled=false;availability();setStatus('busy','Extracting… previous results remain visible.');
 try{const problem=await api.Validate(state,s);if(problem)throw new Error(problem);if(cancelled||rev!==revision){setStatus('stale','Your latest edits are waiting.');return;}const next:Result=await api.Extract(state,s);if(cancelled||rev!==revision){setStatus('stale','Ignored a cancelled or outdated response. Your latest edits are waiting.');return;}result=next;lastRevision=rev;renderResults();setStatus(next.truncated?'warning':'success',next.truncated?'Completed, but the input was truncated. Review before using these values.':'Up to date. Typed output and confidence still require verification.');}
 catch(e){setStatus(cancelled?'ready':'error',cancelled?'Extraction cancelled. Previous results remain available.':message(e));}
 finally{busy=false;availability();if(!cancelled&&auto.checked&&(queued||revision!==rev))schedulePreview();}
}
async function guard(action:()=>Promise<any>|any){try{await action();}catch(e){setStatus(busy?'busy':'error',message(e));}}
async function openSchema(){const file=await api.OpenSchema();if(!file.path)return;schema=file.schema;schemaPath=file.path;selected=schema.fields.length?0:-1;edited();renderFields();}
async function openDocument(){const file=await api.OpenDocument();if(!file.path)return;doc.value=file.text;documentPath=file.path;edited();}
async function saveSchema(){const path=await api.SaveSchema(schema);if(path){schemaPath=path;saveWorkspace();setStatus(busy?'busy':result&&lastRevision===revision?'success':'ready','Schema saved. Draft edits continue to save separately.');}}
function openConnection(){const dialog=$<HTMLDialogElement>('connection-dialog');$<HTMLInputElement>('account').value=connection.account;$<HTMLInputElement>('token').value='';$('connection-state').textContent=connection.message;$('connection-error').textContent='';dialog.showModal();}
async function configure(remember:boolean){connection=await api.ConfigureConnection($<HTMLInputElement>('account').value,$<HTMLInputElement>('token').value,remember);availability();$('connection-state').textContent=connection.message;}
function embedded(){return !!document.querySelector<HTMLInputElement>('input[name=input-mode]:checked')?.value.includes('embed');}
function exportStyle() {
    return $<HTMLSelectElement>('language').value === 'PowerShell'
        ? $<HTMLSelectElement>('powershell-style').value
        : 'standalone';
}
async function updateCode() {
    const rev = ++exportRevision;
    copyCodeFeedback.reset();
    $<HTMLButtonElement>('copy-code').disabled = true;
    $<HTMLButtonElement>('save-code').disabled = true;
    $<HTMLTextAreaElement>('code').value = '';
    $('export-error').textContent = '';
    const isPowerShell = $<HTMLSelectElement>('language').value === 'PowerShell';
    $('powershell-style-control').hidden = !isPowerShell;
    $('powershell-hint').hidden = !isPowerShell;
    try {
        const generated = await api.GenerateCode($<HTMLSelectElement>('language').value, doc.value, schema, embedded(), exportStyle());
        if (rev !== exportRevision) {
            return;
        }
        $<HTMLTextAreaElement>('code').value = generated.code;
        $('export-note').textContent = generated.note;
        $<HTMLButtonElement>('copy-code').disabled = false;
        $<HTMLButtonElement>('save-code').disabled = false;
    } catch (e) {
        if (rev === exportRevision) {
            $('export-error').textContent = message(e);
        }
    }
}
function openExport(){if(currentError(false))return;$<HTMLDialogElement>('export-dialog').showModal();void updateCode();}
const actions:Record<string,()=>any>={new:newSchema,undo:undoNew,'open-schema':openSchema,'save-schema':saveSchema,'open-document':openDocument,'add-field':addField,export:openExport,connect:openConnection,theme:()=>{setTheme(theme==='dark'?'light':'dark');saveWorkspace();}};
for(const [id,action] of Object.entries(actions))$(id==='connect'?'connection':id).onclick=()=>void guard(action);
$('example').onclick=()=>{doc.value=sampleDocument;documentPath='';schema=sampleSchema();schemaPath='';selected=0;edited();renderFields();setStatus('ready','Fictional invoice loaded. Inspect the fields, then preview or export.');};
doc.oninput=()=>{documentPath='';edited();};
fieldName.oninput=()=>{if(schema.fields[selected]){schema.fields[selected].name=fieldName.value;fieldList();edited();}};
fieldName.onfocus=()=>fieldName.select();
instructions.oninput=()=>{if(schema.fields[selected]){schema.fields[selected].instructions=instructions.value;edited();}};
fieldType.onchange=()=>{const f=schema.fields[selected];if(f){f.type=fieldType.value;f.settings=newField(f.name,f.type).settings;edited();renderFields();}};
$('duplicate').onclick=()=>{if(selected<0||schema.fields.length>=500)return;const f=clone(schema.fields[selected]);let i=1;while(schema.fields.some(x=>x.name.toLowerCase()===`copy_${i}`))i++;f.name=`copy_${i}`;schema.fields.splice(selected+1,0,f);selected++;edited();renderFields();};
$('delete').onclick=()=>{if(selected<0)return;schema.fields.splice(selected,1);selected=Math.min(selected,schema.fields.length-1);edited();renderFields();};
function move(delta:number){if(selected<0||selected+delta<0||selected+delta>=schema.fields.length)return;const [f]=schema.fields.splice(selected,1);selected+=delta;schema.fields.splice(selected,0,f);edited();renderFields();}
$('move-up').onclick=()=>move(-1);$('move-down').onclick=()=>move(1);
auto.onchange=()=>{clearTimeout(previewTimer);availability();if(auto.checked&&lastRevision!==revision)schedulePreview();};
$('extract').onclick=()=>void extract();$('cancel').onclick=()=>{cancelled=true;queued=false;clearTimeout(previewTimer);void api.CancelExtraction();};
$('copy-json').onclick = () => void guard(async () => {
    if (result && lastRevision === revision) {
        const copiedRevision = revision;
        const confirmed = await copyJSONFeedback.copy(JSON.stringify(valuesOnly(result, schema.fields), null, 2), text => api.CopyText(text));
        if (confirmed && copiedRevision === revision) {
            setStatus('success', 'Current result values copied as JSON.');
        }
    }
});
for(const button of document.querySelectorAll<HTMLButtonElement>('.close-dialog'))button.onclick=()=>button.closest('dialog')!.close();
$('connection-form').onsubmit=async e=>{e.preventDefault();try{await configure($<HTMLInputElement>('remember').checked);$<HTMLDialogElement>('connection-dialog').close();setStatus('ready','Credentials configured. Choose Extract now to process your document.');}catch(error){$('connection-error').textContent=message(error);}};
$('test-connection').onclick=async()=>{const button=$<HTMLButtonElement>('test-connection');button.disabled=true;try{await configure($<HTMLInputElement>('remember').checked);$('connection-state').textContent='Testing with one small extraction request…';$('connection-state').textContent=await api.TestConnection();$('connection-error').textContent='';}catch(e){$('connection-error').textContent=message(e);}finally{button.disabled=false;}};
$('forget').onclick=async()=>{try{connection=await api.ForgetConnection();availability();$('connection-state').textContent=connection.configured?'Saved token removed. Environment credentials are still available.':'Saved credentials removed.';$<HTMLInputElement>('token').value='';}catch(e){$('connection-error').textContent=message(e);}};
$('setup-help').onclick=()=>{$('setup-instructions').hidden=!$('setup-instructions').hidden;};
$('language').onchange=()=>void updateCode();$('powershell-style').onchange=()=>void updateCode();for(const radio of document.querySelectorAll<HTMLInputElement>('input[name=input-mode]'))radio.onchange=()=>void updateCode();
$('copy-code').onclick = async () => {
    const rev = exportRevision;
    try {
        const confirmed = await copyCodeFeedback.copy($<HTMLTextAreaElement>('code').value, text => api.CopyText(text));
        if (confirmed) {
            $('export-error').textContent = '';
        }
    } catch (e) {
        if (rev === exportRevision) {
            $('export-error').textContent = message(e);
        }
    }
};
$<HTMLDialogElement>('export-dialog').onclose = () => {
    exportRevision++;
    copyCodeFeedback.reset();
};
$('save-code').onclick=async()=>{try{const path=await api.SaveCode($<HTMLSelectElement>('language').value,doc.value,schema,embedded(),exportStyle());if(path){$('export-note').textContent=`Saved ${path}`;}}catch(e){$('export-error').textContent=message(e);}};
document.addEventListener('keydown',e=>{const inDialog=!!document.querySelector('dialog[open]');if(e.altKey&&e.key.toLowerCase()==='n'&&!inDialog){e.preventDefault();if(selected>=0){fieldName.focus();fieldName.select();}return;}if(inDialog)return;if(e.ctrlKey||e.metaKey){const key=e.key.toLowerCase();if(key==='n'){e.preventDefault();newSchema();}else if(key==='o'){e.preventDefault();void guard(openSchema);}else if(key==='f'&&e.shiftKey){e.preventDefault();addField();}}});
function splitter(id:string,vertical:boolean){const node=$(id);let value=vertical?46:60;const apply=(n:number)=>{value=Math.max(vertical?28:35,Math.min(vertical?65:75,n));$(vertical?'top-panes':'workbench').style.setProperty(vertical?'--document-width':'--top-height',value+'%');};node.onpointerdown=e=>{e.preventDefault();node.setPointerCapture(e.pointerId);const move=(m:PointerEvent)=>{const r=$(vertical?'top-panes':'workbench').getBoundingClientRect();apply(100*(vertical?m.clientX-r.left:m.clientY-r.top)/(vertical?r.width:r.height));};node.onpointermove=move;node.onpointerup=()=>{node.onpointermove=null;};};node.onkeydown=e=>{if([vertical?'ArrowLeft':'ArrowUp',vertical?'ArrowRight':'ArrowDown'].includes(e.key)){e.preventDefault();apply(value+(e.key=== (vertical?'ArrowLeft':'ArrowUp')?-2:2));}};}
splitter('column-splitter',true);splitter('row-splitter',false);
async function start(){
 if(!api){setStatus('error','The desktop connection is unavailable. Start ClairicStudio.exe, or use the documented developer preview.');return;}
 try{const initial=await api.Bootstrap();schema=initial.workspace.schema;selected=schema.fields.length?0:-1;documentPath=initial.workspace.document_path||'';schemaPath=initial.workspace.schema_path||'';doc.value=initial.document||'';connection=initial.connection;setTheme(initial.workspace.theme);renderFields();renderResults();setStatus('ready',initial.notice||'Ready.');booting=false;if(previewMode)$('preview-notice').hidden=false;
  if(window.runtime){EventsOn('studio:action',(action:string)=>{if(actions[action])void guard(actions[action]);});EventsOn('studio:save-error',(text:string)=>{setStatus(busy?'busy':'warning',text);});}
  await api.Ready();if(!connection.configured&&!previewMode)openConnection();
 }catch(e){setStatus('error',message(e));}
}
void start();
