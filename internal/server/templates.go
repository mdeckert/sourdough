package server

const completePageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Complete Bake</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 20px;
            padding: 40px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            max-width: 500px;
            width: 100%;
        }
        h1 { color: #333; margin-bottom: 10px; font-size: 28px; text-align: center; }
        .subtitle { color: #666; text-align: center; margin-bottom: 30px; font-size: 14px; }
        .form-group { margin-bottom: 25px; }
        label { display: block; color: #555; margin-bottom: 8px; font-weight: 500; font-size: 14px; }
        .radio-group { display: flex; gap: 10px; flex-wrap: wrap; }
        .radio-option input[type="radio"] { display: none; }
        .radio-option label {
            display: block;
            padding: 12px 16px;
            border: 2px solid #e0e0e0;
            border-radius: 10px;
            text-align: center;
            cursor: pointer;
            transition: all 0.3s;
            font-size: 14px;
        }
        .radio-option input[type="radio"]:checked + label {
            background: #667eea;
            border-color: #667eea;
            color: white;
        }
        textarea {
            width: 100%;
            padding: 16px;
            border: 2px solid #e0e0e0;
            border-radius: 12px;
            font-size: 16px;
            font-family: inherit;
            resize: vertical;
            min-height: 80px;
        }
        textarea:focus { outline: none; border-color: #667eea; }
        button {
            width: 100%;
            padding: 18px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            border-radius: 12px;
            font-size: 18px;
            font-weight: 600;
            cursor: pointer;
            transition: transform 0.2s;
        }
        button:hover { transform: translateY(-2px); }
        button:disabled { background: #ccc; cursor: not-allowed; }
        .success, .error {
            padding: 15px;
            border-radius: 12px;
            text-align: center;
            margin-bottom: 20px;
            display: none;
        }
        .success { background: #10b981; color: white; }
        .error { background: #ef4444; color: white; }
        .slider-container { display: flex; align-items: center; gap: 15px; }
        input[type="range"] { flex: 1; height: 8px; }
        .slider-value { font-size: 24px; font-weight: bold; color: #667eea; min-width: 50px; text-align: center; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🏁 Complete Bake</h1>
        <p class="subtitle">Rate your sourdough</p>

        <div id="success" class="success"></div>
        <div id="error" class="error"></div>

        <form id="assessmentForm">
            <div class="form-group">
                <label>Proof Level</label>
                <div class="radio-group">
                    <div class="radio-option">
                        <input type="radio" id="under" name="proof" value="underproofed" checked>
                        <label for="under">Underproofed</label>
                    </div>
                    <div class="radio-option">
                        <input type="radio" id="good" name="proof" value="good">
                        <label for="good">Good</label>
                    </div>
                    <div class="radio-option">
                        <input type="radio" id="over" name="proof" value="overproofed">
                        <label for="over">Overproofed</label>
                    </div>
                </div>
            </div>

            <div class="form-group">
                <label>Crumb Quality (1-10)</label>
                <div class="slider-container">
                    <input type="range" id="crumb" min="1" max="10" value="5">
                    <div class="slider-value" id="crumbValue">5</div>
                </div>
            </div>

            <div class="form-group">
                <label>Browning</label>
                <div class="radio-group">
                    <div class="radio-option">
                        <input type="radio" id="bnone" name="browning" value="none">
                        <label for="bnone">None</label>
                    </div>
                    <div class="radio-option">
                        <input type="radio" id="bslight" name="browning" value="slight" checked>
                        <label for="bslight">Slight</label>
                    </div>
                    <div class="radio-option">
                        <input type="radio" id="bgood" name="browning" value="good">
                        <label for="bgood">Good</label>
                    </div>
                    <div class="radio-option">
                        <input type="radio" id="bover" name="browning" value="over">
                        <label for="bover">Over</label>
                    </div>
                </div>
            </div>

            <div class="form-group">
                <label>Overall Score (1-10)</label>
                <div class="slider-container">
                    <input type="range" id="score" min="1" max="10" value="5">
                    <div class="slider-value" id="scoreValue">5</div>
                </div>
            </div>

            <div class="form-group">
                <label>Notes (optional)</label>
                <textarea id="notes" placeholder="Any additional observations..."></textarea>
            </div>

            <button type="submit">Complete Bake</button>
        </form>
    </div>

    <script>
        const crumbSlider = document.getElementById('crumb');
        const crumbValue = document.getElementById('crumbValue');
        const scoreSlider = document.getElementById('score');
        const scoreValue = document.getElementById('scoreValue');

        crumbSlider.oninput = () => crumbValue.textContent = crumbSlider.value;
        scoreSlider.oninput = () => scoreValue.textContent = scoreSlider.value;

        document.getElementById('assessmentForm').onsubmit = async (e) => {
            e.preventDefault();

            const data = {
                proof_level: document.querySelector('input[name="proof"]:checked').value,
                crumb_quality: parseInt(crumbSlider.value),
                browning: document.querySelector('input[name="browning"]:checked').value,
                score: parseInt(scoreSlider.value),
                notes: document.getElementById('notes').value
            };

            const button = document.querySelector('button');
            button.disabled = true;
            button.textContent = 'Completing...';

            try {
                const response = await fetch('/log/bake-complete', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ assessment: data })
                });

                if (response.ok) {
                    showSuccess('Bake completed! Great work! 🍞');
                    setTimeout(() => button.textContent = 'Complete Bake', 2000);
                } else {
                    const text = await response.text();
                    showError('Error: ' + text);
                    button.disabled = false;
                    button.textContent = 'Complete Bake';
                }
            } catch (error) {
                showError('Network error: ' + error.message);
                button.disabled = false;
                button.textContent = 'Complete Bake';
            }
        };

        function showSuccess(message) {
            document.getElementById('success').textContent = message;
            document.getElementById('success').style.display = 'block';
            document.getElementById('error').style.display = 'none';
        }

        function showError(message) {
            document.getElementById('error').textContent = message;
            document.getElementById('error').style.display = 'block';
            document.getElementById('success').style.display = 'none';
        }
    </script>
</body>
</html>`

const notesPageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Add Note</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 20px;
            padding: 40px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            max-width: 500px;
            width: 100%;
        }
        h1 {
            color: #333;
            margin-bottom: 10px;
            font-size: 28px;
            text-align: center;
        }
        .subtitle {
            color: #666;
            text-align: center;
            margin-bottom: 30px;
            font-size: 14px;
        }
        .input-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            color: #555;
            margin-bottom: 8px;
            font-weight: 500;
            font-size: 14px;
        }
        textarea {
            width: 100%;
            padding: 16px;
            border: 2px solid #e0e0e0;
            border-radius: 12px;
            font-size: 16px;
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            transition: border-color 0.3s;
            resize: vertical;
            min-height: 150px;
        }
        textarea:focus {
            outline: none;
            border-color: #f5576c;
        }
        button {
            width: 100%;
            padding: 18px;
            background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
            color: white;
            border: none;
            border-radius: 12px;
            font-size: 18px;
            font-weight: 600;
            cursor: pointer;
            transition: transform 0.2s, box-shadow 0.2s;
            box-shadow: 0 4px 15px rgba(245, 87, 108, 0.4);
        }
        button:hover {
            transform: translateY(-2px);
            box-shadow: 0 6px 20px rgba(245, 87, 108, 0.6);
        }
        button:active {
            transform: translateY(0);
        }
        button:disabled {
            background: #ccc;
            cursor: not-allowed;
            box-shadow: none;
        }
        .success {
            background: #10b981;
            color: white;
            padding: 15px;
            border-radius: 12px;
            text-align: center;
            margin-bottom: 20px;
            display: none;
            animation: slideIn 0.3s;
        }
        .error {
            background: #ef4444;
            color: white;
            padding: 15px;
            border-radius: 12px;
            text-align: center;
            margin-bottom: 20px;
            display: none;
        }
        @keyframes slideIn {
            from {
                opacity: 0;
                transform: translateY(-10px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }
        .quick-notes {
            display: flex;
            flex-wrap: wrap;
            gap: 8px;
            margin-bottom: 20px;
        }
        .quick-note {
            padding: 8px 12px;
            background: #f3f4f6;
            border: 2px solid transparent;
            border-radius: 8px;
            cursor: pointer;
            font-size: 14px;
            transition: all 0.2s;
        }
        .quick-note:hover {
            background: #e5e7eb;
            border-color: #f5576c;
        }
        .char-count {
            text-align: right;
            color: #999;
            font-size: 12px;
            margin-top: 5px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>📝 Add Note</h1>
        <p class="subtitle">Sourdough Logger</p>

        <div id="success" class="success"></div>
        <div id="error" class="error"></div>

        <div class="input-group">
            <label>Quick Phrases</label>
            <div class="quick-notes">
                <div class="quick-note" onclick="appendNote('Good oven spring')">Good oven spring</div>
                <div class="quick-note" onclick="appendNote('Underproofed')">Underproofed</div>
                <div class="quick-note" onclick="appendNote('Overproofed')">Overproofed</div>
                <div class="quick-note" onclick="appendNote('Dense crumb')">Dense crumb</div>
                <div class="quick-note" onclick="appendNote('Open crumb')">Open crumb</div>
                <div class="quick-note" onclick="appendNote('Great flavor')">Great flavor</div>
                <div class="quick-note" onclick="appendNote('Too sour')">Too sour</div>
                <div class="quick-note" onclick="appendNote('Not sour enough')">Not sour enough</div>
                <div class="quick-note" onclick="appendNote('Crust too dark')">Crust too dark</div>
                <div class="quick-note" onclick="appendNote('Perfect crust')">Perfect crust</div>
            </div>
        </div>

        <div class="input-group">
            <label for="note">Your Note</label>
            <textarea id="note" placeholder="Enter observations, tasting notes, crumb analysis, etc..." autofocus></textarea>
            <div class="char-count"><span id="count">0</span> characters</div>
        </div>

        <button onclick="addNote()">Add Note</button>
    </div>

    <script>
        const textarea = document.getElementById('note');
        const countEl = document.getElementById('count');

        textarea.addEventListener('input', function() {
            countEl.textContent = this.value.length;
        });

        function appendNote(text) {
            const current = textarea.value;
            if (current && !current.endsWith('. ') && !current.endsWith('.\n')) {
                textarea.value = current + '. ' + text;
            } else {
                textarea.value = current + text;
            }
            countEl.textContent = textarea.value.length;
            textarea.focus();
        }

        async function addNote() {
            const note = document.getElementById('note').value.trim();

            if (!note) {
                showError('Please enter a note');
                return;
            }

            const button = document.querySelector('button');
            button.disabled = true;
            button.textContent = 'Adding...';

            try {
                const response = await fetch('/log/note', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ note: note })
                });

                if (response.ok) {
                    showSuccess('Note added successfully!');
                    document.getElementById('note').value = '';
                    countEl.textContent = '0';

                    setTimeout(() => {
                        button.disabled = false;
                        button.textContent = 'Add Note';
                    }, 1000);
                } else {
                    const text = await response.text();
                    showError('Error: ' + text);
                    button.disabled = false;
                    button.textContent = 'Add Note';
                }
            } catch (error) {
                showError('Network error: ' + error.message);
                button.disabled = false;
                button.textContent = 'Add Note';
            }
        }

        function showSuccess(message) {
            const el = document.getElementById('success');
            el.textContent = message;
            el.style.display = 'block';
            document.getElementById('error').style.display = 'none';

            setTimeout(() => {
                el.style.display = 'none';
            }, 3000);
        }

        function showError(message) {
            const el = document.getElementById('error');
            el.textContent = message;
            el.style.display = 'block';
            document.getElementById('success').style.display = 'none';
        }
    </script>
</body>
</html>`

const tempPageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Log Temperature</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 20px;
            padding: 40px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            max-width: 400px;
            width: 100%;
        }
        h1 {
            color: #333;
            margin-bottom: 10px;
            font-size: 28px;
            text-align: center;
        }
        .subtitle {
            color: #666;
            text-align: center;
            margin-bottom: 30px;
            font-size: 14px;
        }
        .input-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            color: #555;
            margin-bottom: 8px;
            font-weight: 500;
            font-size: 14px;
        }
        input[type="number"] {
            width: 100%;
            padding: 16px;
            border: 2px solid #e0e0e0;
            border-radius: 12px;
            font-size: 24px;
            text-align: center;
            transition: border-color 0.3s;
            font-weight: 600;
        }
        input[type="number"]:focus {
            outline: none;
            border-color: #667eea;
        }
        .temp-unit {
            text-align: center;
            color: #999;
            font-size: 18px;
            margin-top: -10px;
            margin-bottom: 20px;
        }
        .radio-group {
            display: flex;
            gap: 15px;
            justify-content: center;
            margin-bottom: 25px;
        }
        .radio-option {
            flex: 1;
        }
        .radio-option input[type="radio"] {
            display: none;
        }
        .radio-option label {
            display: block;
            padding: 12px;
            border: 2px solid #e0e0e0;
            border-radius: 10px;
            text-align: center;
            cursor: pointer;
            transition: all 0.3s;
            font-weight: 500;
        }
        .radio-option input[type="radio"]:checked + label {
            background: #667eea;
            border-color: #667eea;
            color: white;
        }
        button {
            width: 100%;
            padding: 18px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            border-radius: 12px;
            font-size: 18px;
            font-weight: 600;
            cursor: pointer;
            transition: transform 0.2s, box-shadow 0.2s;
            box-shadow: 0 4px 15px rgba(102, 126, 234, 0.4);
        }
        button:hover {
            transform: translateY(-2px);
            box-shadow: 0 6px 20px rgba(102, 126, 234, 0.6);
        }
        button:active {
            transform: translateY(0);
        }
        button:disabled {
            background: #ccc;
            cursor: not-allowed;
            box-shadow: none;
        }
        .success {
            background: #10b981;
            color: white;
            padding: 15px;
            border-radius: 12px;
            text-align: center;
            margin-bottom: 20px;
            display: none;
            animation: slideIn 0.3s;
        }
        .error {
            background: #ef4444;
            color: white;
            padding: 15px;
            border-radius: 12px;
            text-align: center;
            margin-bottom: 20px;
            display: none;
        }
        @keyframes slideIn {
            from {
                opacity: 0;
                transform: translateY(-10px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }
        .quick-buttons {
            display: grid;
            grid-template-columns: repeat(3, 1fr);
            gap: 8px;
            margin-bottom: 20px;
        }
        .quick-temp {
            padding: 10px;
            background: #f3f4f6;
            border: 2px solid transparent;
            border-radius: 8px;
            cursor: pointer;
            text-align: center;
            font-weight: 500;
            transition: all 0.2s;
        }
        .quick-temp:hover {
            background: #e5e7eb;
            border-color: #667eea;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🌡️ Log Temperature</h1>
        <p class="subtitle">Sourdough Logger</p>

        <div id="success" class="success"></div>
        <div id="error" class="error"></div>

        <div class="input-group">
            <label>Quick Select</label>
            <div class="quick-buttons">
                <div class="quick-temp" onclick="setTemp(68)">68°F</div>
                <div class="quick-temp" onclick="setTemp(70)">70°F</div>
                <div class="quick-temp" onclick="setTemp(72)">72°F</div>
                <div class="quick-temp" onclick="setTemp(74)">74°F</div>
                <div class="quick-temp" onclick="setTemp(76)">76°F</div>
                <div class="quick-temp" onclick="setTemp(78)">78°F</div>
            </div>
        </div>

        <div class="input-group">
            <label for="temp">Temperature</label>
            <input type="number" id="temp" min="60" max="80" step="1" placeholder="76" autofocus>
            <div class="temp-unit">°F</div>
        </div>

        <div class="radio-group">
            <div class="radio-option">
                <input type="radio" id="kitchen" name="tempType" value="kitchen" checked>
                <label for="kitchen">Kitchen</label>
            </div>
            <div class="radio-option">
                <input type="radio" id="dough" name="tempType" value="dough">
                <label for="dough">Dough</label>
            </div>
        </div>

        <button onclick="logTemp()">Log Temperature</button>
    </div>

    <script>
        function setTemp(value) {
            document.getElementById('temp').value = value;
        }

        async function logTemp() {
            const temp = document.getElementById('temp').value;
            const tempType = document.querySelector('input[name="tempType"]:checked').value;

            if (!temp || temp < 60 || temp > 80) {
                showError('Please enter a temperature between 60°F and 80°F');
                return;
            }

            const button = document.querySelector('button');
            button.disabled = true;
            button.textContent = 'Logging...';

            try {
                const url = '/log/temp/' + temp + (tempType === 'dough' ? '?type=dough' : '');
                const response = await fetch(url, { method: 'POST' });

                if (response.ok) {
                    showSuccess(temp + '°F logged successfully!');
                    document.getElementById('temp').value = '';

                    // Re-enable after 1 second
                    setTimeout(() => {
                        button.disabled = false;
                        button.textContent = 'Log Temperature';
                    }, 1000);
                } else {
                    const text = await response.text();
                    showError('Error: ' + text);
                    button.disabled = false;
                    button.textContent = 'Log Temperature';
                }
            } catch (error) {
                showError('Network error: ' + error.message);
                button.disabled = false;
                button.textContent = 'Log Temperature';
            }
        }

        function showSuccess(message) {
            const el = document.getElementById('success');
            el.textContent = message;
            el.style.display = 'block';
            document.getElementById('error').style.display = 'none';

            setTimeout(() => {
                el.style.display = 'none';
            }, 3000);
        }

        function showError(message) {
            const el = document.getElementById('error');
            el.textContent = message;
            el.style.display = 'block';
            document.getElementById('success').style.display = 'none';
        }

        // Allow Enter key to submit
        document.getElementById('temp').addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                logTemp();
            }
        });
    </script>
</body>
</html>`
