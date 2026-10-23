type CopyButton = Pick<HTMLButtonElement, 'textContent' | 'setAttribute'>;

export function createCopyFeedback(button: CopyButton, duration = 2000) {
    const label = button.textContent;
    let revision = 0;
    let timer: ReturnType<typeof setTimeout> | undefined;
    button.setAttribute('aria-live', 'polite');
    button.setAttribute('aria-atomic', 'true');

    function reset() {
        revision++;
        clearTimeout(timer);
        button.textContent = label;
    }

    async function copy(text: string, write: (text: string) => Promise<void>) {
        reset();
        const current = revision;
        await write(text);
        if (current !== revision) {
            return false;
        }
        button.textContent = '✓ Copied';
        timer = setTimeout(reset, duration);
        return true;
    }

    return { copy, reset };
}
