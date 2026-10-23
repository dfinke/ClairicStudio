import test from 'node:test';
import assert from 'node:assert/strict';
import { createCopyFeedback } from '../.test/clipboard.js';

function button(label = 'Copy code') {
    return { textContent: label, setAttribute() {} };
}

test('confirmation waits for clipboard success and resets after two seconds', async t => {
    t.mock.timers.enable({ apis: ['setTimeout'] });
    const control = button();
    const feedback = createCopyFeedback(control);
    let finish;
    const copied = feedback.copy('the program', text => {
        assert.equal(text, 'the program');
        return new Promise(resolve => { finish = resolve; });
    });
    assert.equal(control.textContent, 'Copy code');
    finish();
    assert.equal(await copied, true);
    assert.equal(control.textContent, '✓ Copied');
    t.mock.timers.tick(1999);
    assert.equal(control.textContent, '✓ Copied');
    t.mock.timers.tick(1);
    assert.equal(control.textContent, 'Copy code');
});

test('clipboard failures do not claim a successful copy', async t => {
    t.mock.timers.enable({ apis: ['setTimeout'] });
    const control = button('Copy JSON');
    const feedback = createCopyFeedback(control);
    await feedback.copy('{}', async () => {});
    assert.equal(control.textContent, '✓ Copied');
    await assert.rejects(feedback.copy('{}', async () => { throw new Error('Clipboard unavailable'); }), /Clipboard unavailable/);
    assert.equal(control.textContent, 'Copy JSON');
});

test('changing content invalidates an unfinished copy and its confirmation', async () => {
    const control = button();
    const feedback = createCopyFeedback(control);
    let finish;
    const pending = feedback.copy('old program', () => new Promise(resolve => { finish = resolve; }));
    feedback.reset();
    finish();
    assert.equal(await pending, false);
    assert.equal(control.textContent, 'Copy code');
});

test('repeated copies keep the most recent confirmation visible for two seconds', async t => {
    t.mock.timers.enable({ apis: ['setTimeout'] });
    const control = button();
    const feedback = createCopyFeedback(control);
    await feedback.copy('first', async () => {});
    t.mock.timers.tick(1000);
    await feedback.copy('second', async () => {});
    t.mock.timers.tick(1000);
    assert.equal(control.textContent, '✓ Copied');
    t.mock.timers.tick(1000);
    assert.equal(control.textContent, 'Copy code');
});
