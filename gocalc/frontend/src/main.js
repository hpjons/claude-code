const historyEl = document.getElementById('history');
const inputEl = document.getElementById('expr-input');

// Typing one of these into an empty input prefixes it with "ans", so
// "* 2" continues from the previous result. '-' is deliberately excluded:
// a leading minus almost always starts a negative number (e.g. "-5 + 3"),
// and "ans-…" can still be typed explicitly.
const ANS_TRIGGER_OPERATORS = ['+', '*', '/', '^', '%'];

let commandHistory = [];
let historyPos = 0;
let draft = '';
let hasAns = false;

function appendEntry(res) {
    const entry = document.createElement('div');
    entry.className = 'entry';

    const exprLine = document.createElement('div');
    exprLine.className = 'entry-expr';
    exprLine.textContent = '> ' + res.input;

    const resultLine = document.createElement('div');
    resultLine.className = res.isError ? 'entry-result entry-error' : 'entry-result';
    resultLine.textContent = res.isError ? res.errorMsg : res.result;

    entry.appendChild(exprLine);
    entry.appendChild(resultLine);
    historyEl.appendChild(entry);
    historyEl.scrollTop = historyEl.scrollHeight;

    if (!res.isError) {
        hasAns = true;
    }
}

function submit() {
    const text = inputEl.value.trim();
    if (text === '') return;

    commandHistory.push(text);
    historyPos = commandHistory.length;
    draft = '';
    inputEl.value = '';

    if (text.toLowerCase() === 'clear') {
        historyEl.innerHTML = '';
        return;
    }

    window.go.main.App.Evaluate(text)
        .then((res) => appendEntry(res))
        .catch((err) => appendEntry({ input: text, isError: true, errorMsg: String(err) }));
}

function navigateHistory(delta) {
    if (commandHistory.length === 0) return;

    if (historyPos === commandHistory.length) {
        draft = inputEl.value;
    }

    const newPos = Math.max(0, Math.min(commandHistory.length, historyPos + delta));
    historyPos = newPos;
    inputEl.value = historyPos === commandHistory.length ? draft : commandHistory[historyPos];

    requestAnimationFrame(() => {
        inputEl.setSelectionRange(inputEl.value.length, inputEl.value.length);
    });
}

inputEl.addEventListener('keydown', (e) => {
    switch (e.key) {
        case 'Enter':
            submit();
            break;
        case 'ArrowUp':
            e.preventDefault();
            navigateHistory(-1);
            break;
        case 'ArrowDown':
            e.preventDefault();
            navigateHistory(1);
            break;
        case 'Escape':
            inputEl.value = '';
            historyPos = commandHistory.length;
            draft = '';
            break;
    }
});

inputEl.addEventListener('input', () => {
    if (hasAns && inputEl.value.length === 1 && ANS_TRIGGER_OPERATORS.includes(inputEl.value)) {
        inputEl.value = 'ans' + inputEl.value;
        requestAnimationFrame(() => {
            inputEl.setSelectionRange(inputEl.value.length, inputEl.value.length);
        });
    }
});

document.addEventListener('click', () => {
    if (window.getSelection().toString() === '') {
        inputEl.focus();
    }
});

window.go.main.App.InputPlaceholder()
    .then((text) => {
        inputEl.placeholder = text;
    })
    .catch(() => {});

inputEl.focus();
