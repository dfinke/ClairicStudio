import test from 'node:test';
import assert from 'node:assert/strict';
import { types,newField,settingsFor,verifiedSpan,localError,valuesOnly,displayValue } from '../.test/logic.js';

test('source spans account for Unicode without highlighting guesses',()=>{
  const text='🧾 From Northwind';
  assert.deepEqual(verifiedSpan(text,{value:'Northwind',span:[7,16]}),[8,17]);
  assert.deepEqual(verifiedSpan(text,{value:'Northwind',span:[8,17]}),[8,17]);
  assert.equal(verifiedSpan(text,{value:'Invented',span:[8,17]}),null);
  assert.equal(verifiedSpan(text,{value:null,span:[0,2]}),null);
  assert.equal(verifiedSpan(text,{value:'Northwind',span:[-1,8]}),null);
});
test('all nine types expose their relevant settings',()=>{
  assert.equal(types.length,9);
  for(const type of types)assert.ok(settingsFor(newField('value',type)).length);
  assert.ok(settingsFor(newField('value','forecast')).includes('horizon'));
  assert.ok(settingsFor({...newField('value','list'),settings:{items:'number'}}).includes('decimals'));
});
test('draft validation permits export without document and rejects ambiguous names',()=>{
  const schema={version:1,fields:[newField('vendor')]};
  assert.equal(localError(schema,'',false),'');
  assert.ok(localError(schema,''));
  schema.fields.push(newField('VENDOR'));
  assert.match(localError(schema,'invoice'),/unique/);
  schema.fields=[newField('bad name')];
  assert.match(localError(schema,'invoice'),/Field names/);
});
test('JSON copying retains typed values, null and safe property names',()=>{
  const answers=JSON.parse('{"amount":{"value":23.76},"items":{"value":["A","B"]},"missing":{"value":null},"__proto__":{"value":"safe"}}');
  const values=valuesOnly({answers,truncated:false},Object.keys(answers).map(n=>newField(n)));
  assert.equal(values.amount,23.76);
  assert.deepEqual(values.items,['A','B']);
  assert.equal(values.missing,null);
  assert.equal(values.__proto__,'safe');
  assert.equal(Object.getPrototypeOf(values),null);
  assert.equal(displayValue(null),'Not found');
  assert.throws(()=>valuesOnly({answers:{},truncated:false},[newField('missing')]),/Missing answer/);
});
