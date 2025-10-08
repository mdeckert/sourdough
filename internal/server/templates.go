package server

// navDropdownHTML is the navigation dropdown added to all UI pages
const navDropdownHTML = `
<div style="margin-top: 30px; padding-top: 20px; border-top: 2px solid #e0e0e0;">
    <label style="display: block; color: #666; margin-bottom: 10px; font-weight: 500; font-size: 14px; text-align: center;">Quick Navigation</label>
    <select onchange="if(this.value) window.location.href=this.value" style="width: 100%; padding: 12px; border: 2px solid #e0e0e0; border-radius: 12px; font-size: 16px; background: white; cursor: pointer;">
        <option value="">Go to...</option>
        <optgroup label="Workflow Events">
            <option value="/loaf/start">🥖 Start Loaf</option>
            <option value="/log/fed">🍞 Fed</option>
            <option value="/log/levain-ready">⏰ Levain Ready</option>
            <option value="/log/mixed">🥣 Mixed</option>
            <option value="/log/fold">🙌 Fold</option>
            <option value="/log/shaped">👐 Shaped</option>
            <option value="/log/fridge-in">❄️ Fridge In</option>
            <option value="/log/oven-in">🔥 Oven In</option>
            <option value="/log/remove-lid">🌡️ Remove Lid</option>
            <option value="/log/oven-out">✅ Oven Out</option>
            <option value="/complete">🎉 Complete</option>
        </optgroup>
        <optgroup label="Logging">
            <option value="/temp">🌡️ Log Temperature</option>
            <option value="/notes">📝 Add Note</option>
        </optgroup>
        <optgroup label="View">
            <option value="/view/status">📊 View Status</option>
            <option value="/view/history">📚 View History</option>
            <option value="/qrcodes.pdf">📱 Get QR Codes</option>
        </optgroup>
    </select>
</div>
`

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

            ` + navDropdownHTML + `
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
                const response = await fetch('/log/loaf-complete', {
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
        .image-upload {
            margin-top: 15px;
        }
        .image-upload-btn {
            display: inline-block;
            padding: 12px 24px;
            background: #3b82f6;
            color: white;
            border-radius: 12px;
            cursor: pointer;
            font-size: 16px;
            transition: all 0.2s;
            border: none;
        }
        .image-upload-btn:hover {
            background: #2563eb;
            transform: translateY(-2px);
        }
        .image-preview {
            margin-top: 15px;
            display: none;
        }
        .image-preview img {
            max-width: 100%;
            border-radius: 12px;
            box-shadow: 0 4px 15px rgba(0,0,0,0.2);
        }
        .image-preview-actions {
            margin-top: 10px;
            display: flex;
            gap: 10px;
        }
        .remove-image-btn {
            padding: 8px 16px;
            background: #ef4444;
            color: white;
            border: none;
            border-radius: 8px;
            cursor: pointer;
            font-size: 14px;
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
            <label>Dough Feel</label>
            <div class="quick-notes">
                <div class="quick-note" onclick="appendNote('Sticky')">Sticky</div>
                <div class="quick-note" onclick="appendNote('Slack')">Slack</div>
                <div class="quick-note" onclick="appendNote('Tight')">Tight</div>
                <div class="quick-note" onclick="appendNote('Jiggly')">Jiggly</div>
                <div class="quick-note" onclick="appendNote('Smooth')">Smooth</div>
            </div>
        </div>

        <div class="input-group">
            <label for="doughTemp">Dough Temperature (optional)</label>
            <input type="number" id="doughTemp" placeholder="e.g., 76" step="1" style="width: 100%; padding: 16px; border: 2px solid #e0e0e0; border-radius: 12px; font-size: 18px; text-align: center;">
        </div>

        <div class="input-group">
            <label for="note">Your Note</label>
            <textarea id="note" placeholder="Enter observations about dough, levain, proofing progress..." autofocus></textarea>
            <div class="char-count"><span id="count">0</span> characters</div>
        </div>

        <div class="input-group image-upload">
            <label>📷 Add Photo (optional)</label>
            <input type="file" id="imageInput" accept="image/*" capture="environment" style="display: none" onchange="handleImageSelect(event)">
            <button class="image-upload-btn" onclick="document.getElementById('imageInput').click()">
                Take Photo / Choose Image
            </button>
            <div id="imagePreview" class="image-preview">
                <img id="previewImg" src="" alt="Preview">
                <div class="image-preview-actions">
                    <button class="remove-image-btn" onclick="removeImage()">Remove Image</button>
                </div>
            </div>
        </div>

        <button onclick="addNote()">Add Note</button>

        ` + navDropdownHTML + `
    </div>

    <script>
        const textarea = document.getElementById('note');
        const countEl = document.getElementById('count');
        let selectedImage = null;

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

        function handleImageSelect(event) {
            const file = event.target.files[0];
            if (!file) return;

            // Validate file type
            if (!file.type.startsWith('image/')) {
                showError('Please select an image file');
                return;
            }

            // Compress and resize image before upload
            compressImage(file, 1920, 0.85).then(compressedBlob => {
                selectedImage = compressedBlob;

                // Show preview
                const reader = new FileReader();
                reader.onload = function(e) {
                    document.getElementById('previewImg').src = e.target.result;
                    document.getElementById('imagePreview').style.display = 'block';
                };
                reader.readAsDataURL(compressedBlob);
            }).catch(err => {
                showError('Failed to process image: ' + err.message);
            });
        }

        function compressImage(file, maxWidth, quality) {
            return new Promise((resolve, reject) => {
                const reader = new FileReader();
                reader.onload = function(e) {
                    const img = new Image();
                    img.onload = function() {
                        // Calculate new dimensions
                        let width = img.width;
                        let height = img.height;

                        if (width > maxWidth) {
                            height = (height * maxWidth) / width;
                            width = maxWidth;
                        }

                        // Create canvas and draw resized image
                        const canvas = document.createElement('canvas');
                        canvas.width = width;
                        canvas.height = height;
                        const ctx = canvas.getContext('2d');
                        ctx.drawImage(img, 0, 0, width, height);

                        // Convert to blob with compression
                        canvas.toBlob(
                            blob => {
                                if (!blob) {
                                    reject(new Error('Failed to compress image'));
                                    return;
                                }
                                resolve(blob);
                            },
                            'image/jpeg',
                            quality
                        );
                    };
                    img.onerror = () => reject(new Error('Failed to load image'));
                    img.src = e.target.result;
                };
                reader.onerror = () => reject(new Error('Failed to read file'));
                reader.readAsDataURL(file);
            });
        }

        function removeImage() {
            selectedImage = null;
            document.getElementById('imageInput').value = '';
            document.getElementById('imagePreview').style.display = 'none';
        }

        async function addNote() {
            const note = document.getElementById('note').value.trim();
            const doughTemp = document.getElementById('doughTemp').value.trim();

            if (!note && !selectedImage) {
                showError('Please enter a note or attach an image');
                return;
            }

            const button = document.querySelector('button[onclick="addNote()"]');
            button.disabled = true;
            button.textContent = 'Adding...';

            try {
                // Use FormData to send note, dough temp, and image
                const formData = new FormData();
                formData.append('note', note);
                if (doughTemp) {
                    formData.append('dough_temp', doughTemp);
                }
                if (selectedImage) {
                    formData.append('image', selectedImage);
                }

                const response = await fetch('/log/note', {
                    method: 'POST',
                    body: formData
                });

                if (response.ok) {
                    showSuccess('Note added successfully!');
                    document.getElementById('note').value = '';
                    document.getElementById('doughTemp').value = '';
                    countEl.textContent = '0';
                    removeImage();

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
        .slider-container {
            display: flex;
            align-items: center;
            gap: 15px;
            margin-bottom: 20px;
        }
        input[type="range"] {
            flex: 1;
            height: 8px;
            -webkit-appearance: none;
            appearance: none;
            background: #e0e0e0;
            border-radius: 5px;
            outline: none;
        }
        input[type="range"]::-webkit-slider-thumb {
            -webkit-appearance: none;
            appearance: none;
            width: 24px;
            height: 24px;
            background: #667eea;
            border-radius: 50%;
            cursor: pointer;
        }
        input[type="range"]::-moz-range-thumb {
            width: 24px;
            height: 24px;
            background: #667eea;
            border-radius: 50%;
            cursor: pointer;
            border: none;
        }
        .slider-value {
            font-size: 32px;
            font-weight: bold;
            color: #667eea;
            min-width: 70px;
            text-align: center;
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
            <label>Temperature (60-80°F)</label>
            <div class="slider-container">
                <input type="range" id="tempSlider" min="60" max="80" value="70" step="1">
                <div class="slider-value" id="sliderValue">70°F</div>
            </div>
        </div>

        <div class="input-group">
            <label for="temp">Manual Entry</label>
            <input type="number" id="temp" min="60" max="80" step="1" placeholder="76">
        </div>

        <div class="radio-group">
            <div class="radio-option">
                <input type="radio" id="dough" name="tempType" value="dough">
                <label for="dough" onclick="submitTemp('dough')">Dough</label>
            </div>
            <div class="radio-option">
                <input type="radio" id="oven" name="tempType" value="oven">
                <label for="oven" onclick="submitTemp('oven')">Oven</label>
            </div>
            <div class="radio-option">
                <input type="radio" id="loaf" name="tempType" value="loaf">
                <label for="loaf" onclick="submitTemp('loaf')">Loaf</label>
            </div>
        </div>

        ` + navDropdownHTML + `
    </div>

    <script>
        const slider = document.getElementById('tempSlider');
        const sliderValue = document.getElementById('sliderValue');
        const manualInput = document.getElementById('temp');

        // Update slider display
        slider.oninput = function() {
            sliderValue.textContent = this.value + '°F';
            manualInput.value = ''; // Clear manual input when slider moves
        };

        // Sync manual input to slider (only if within slider range)
        manualInput.oninput = function() {
            if (this.value >= 60 && this.value <= 80) {
                slider.value = this.value;
                sliderValue.textContent = this.value + '°F';
            }
        };

        async function submitTemp(tempType) {
            let temp = manualInput.value || slider.value;

            // Validate temperature is reasonable (0-600°F covers all use cases)
            if (!temp || temp < 0 || temp > 600) {
                showError('Temperature must be between 0°F and 600°F');
                return;
            }

            try {
                // Map tempType to query parameter
                let typeParam = '';
                if (tempType === 'dough' || tempType === 'loaf') {
                    typeParam = '?type=dough';  // Both use dough_temp_f field
                } else if (tempType === 'oven') {
                    typeParam = '?type=oven';
                }
                // No param for kitchen (auto-logged via Ecobee anyway)

                const url = '/log/temp/' + temp + typeParam;
                const response = await fetch(url, { method: 'POST' });

                if (response.ok) {
                    showSuccess(temp + '°F (' + tempType + ') logged!');
                    manualInput.value = '';
                    slider.value = 70;
                    sliderValue.textContent = '70°F';
                } else {
                    const text = await response.text();
                    showError('Error: ' + text);
                }
            } catch (error) {
                showError('Network error: ' + error.message);
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

const ovenInPageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Oven In</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
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
            max-width: 400px;
            width: 100%;
        }
        h1 { color: #333; margin-bottom: 10px; font-size: 28px; text-align: center; }
        .subtitle { color: #666; text-align: center; margin-bottom: 30px; font-size: 14px; }
        .temp-grid {
            display: grid;
            grid-template-columns: repeat(3, 1fr);
            gap: 10px;
            margin-bottom: 20px;
        }
        .temp-btn {
            padding: 20px 10px;
            background: white;
            border: 2px solid #e0e0e0;
            border-radius: 12px;
            font-size: 20px;
            font-weight: 600;
            color: #333;
            cursor: pointer;
            transition: all 0.2s;
        }
        .temp-btn:hover {
            border-color: #f5576c;
            background: #fff5f7;
            transform: translateY(-2px);
        }
        .temp-btn:active { transform: translateY(0); }
        .success, .error {
            padding: 15px;
            border-radius: 12px;
            text-align: center;
            margin-bottom: 20px;
            display: none;
        }
        .success { background: #10b981; color: white; }
        .error { background: #ef4444; color: white; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔥 Oven In</h1>
        <p class="subtitle">Select Oven Temperature</p>
        <div id="success" class="success"></div>
        <div id="error" class="error"></div>
        <div class="temp-grid">
            <button class="temp-btn" onclick="logOvenIn(420)">420°F</button>
            <button class="temp-btn" onclick="logOvenIn(425)">425°F</button>
            <button class="temp-btn" onclick="logOvenIn(430)">430°F</button>
            <button class="temp-btn" onclick="logOvenIn(435)">435°F</button>
            <button class="temp-btn" onclick="logOvenIn(440)">440°F</button>
            <button class="temp-btn" onclick="logOvenIn(445)">445°F</button>
            <button class="temp-btn" onclick="logOvenIn(450)">450°F</button>
            <button class="temp-btn" onclick="logOvenIn(455)">455°F</button>
            <button class="temp-btn" onclick="logOvenIn(460)">460°F</button>
            <button class="temp-btn" onclick="logOvenIn(465)">465°F</button>
            <button class="temp-btn" onclick="logOvenIn(470)">470°F</button>
            <button class="temp-btn" onclick="logOvenIn(475)">475°F</button>
            <button class="temp-btn" onclick="logOvenIn(480)">480°F</button>
        </div>

        ` + navDropdownHTML + `
    </div>
    <script>
        async function logOvenIn(temp) {
            try {
                const response = await fetch('/log/oven-in?temp=' + temp, { method: 'POST' });
                if (response.ok) {
                    document.getElementById('success').textContent = 'Oven In logged at ' + temp + '°F!';
                    document.getElementById('success').style.display = 'block';
                    setTimeout(() => { document.getElementById('success').style.display = 'none'; }, 3000);
                } else {
                    const text = await response.text();
                    document.getElementById('error').textContent = 'Error: ' + text;
                    document.getElementById('error').style.display = 'block';
                }
            } catch (error) {
                document.getElementById('error').textContent = 'Network error: ' + error.message;
                document.getElementById('error').style.display = 'block';
            }
        }
    </script>
</body>
</html>`

const removeLidPageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Remove Lid</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
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
            max-width: 400px;
            width: 100%;
        }
        h1 { color: #333; margin-bottom: 10px; font-size: 28px; text-align: center; }
        .subtitle { color: #666; text-align: center; margin-bottom: 30px; font-size: 14px; }
        .temp-grid {
            display: grid;
            grid-template-columns: repeat(3, 1fr);
            gap: 10px;
            margin-bottom: 20px;
        }
        .temp-btn {
            padding: 20px 10px;
            background: white;
            border: 2px solid #e0e0e0;
            border-radius: 12px;
            font-size: 20px;
            font-weight: 600;
            color: #333;
            cursor: pointer;
            transition: all 0.2s;
        }
        .temp-btn:hover {
            border-color: #f5576c;
            background: #fff5f7;
            transform: translateY(-2px);
        }
        .temp-btn:active { transform: translateY(0); }
        .success, .error {
            padding: 15px;
            border-radius: 12px;
            text-align: center;
            margin-bottom: 20px;
            display: none;
        }
        .success { background: #10b981; color: white; }
        .error { background: #ef4444; color: white; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🌡️ Remove Lid</h1>
        <p class="subtitle">Select Oven Temperature</p>
        <div id="success" class="success"></div>
        <div id="error" class="error"></div>
        <div class="temp-grid">
            <button class="temp-btn" onclick="logRemoveLid(420)">420°F</button>
            <button class="temp-btn" onclick="logRemoveLid(425)">425°F</button>
            <button class="temp-btn" onclick="logRemoveLid(430)">430°F</button>
            <button class="temp-btn" onclick="logRemoveLid(435)">435°F</button>
            <button class="temp-btn" onclick="logRemoveLid(440)">440°F</button>
            <button class="temp-btn" onclick="logRemoveLid(445)">445°F</button>
            <button class="temp-btn" onclick="logRemoveLid(450)">450°F</button>
            <button class="temp-btn" onclick="logRemoveLid(455)">455°F</button>
            <button class="temp-btn" onclick="logRemoveLid(460)">460°F</button>
            <button class="temp-btn" onclick="logRemoveLid(465)">465°F</button>
            <button class="temp-btn" onclick="logRemoveLid(470)">470°F</button>
            <button class="temp-btn" onclick="logRemoveLid(475)">475°F</button>
            <button class="temp-btn" onclick="logRemoveLid(480)">480°F</button>
        </div>

        ` + navDropdownHTML + `
    </div>
    <script>
        async function logRemoveLid(temp) {
            try {
                const response = await fetch('/log/remove-lid?temp=' + temp, { method: 'POST' });
                if (response.ok) {
                    document.getElementById('success').textContent = 'Remove Lid logged at ' + temp + '°F!';
                    document.getElementById('success').style.display = 'block';
                    setTimeout(() => { document.getElementById('success').style.display = 'none'; }, 3000);
                } else {
                    const text = await response.text();
                    document.getElementById('error').textContent = 'Error: ' + text;
                    document.getElementById('error').style.display = 'block';
                }
            } catch (error) {
                document.getElementById('error').textContent = 'Network error: ' + error.message;
                document.getElementById('error').style.display = 'block';
            }
        }
    </script>
</body>
</html>`

const historyViewPageHTML = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Bake History</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            margin-bottom: 20px;
        }
        .header {
            padding: 30px;
            border-bottom: 2px solid #f3f4f6;
            display: flex;
            justify-content: space-between;
            align-items: center;
            flex-wrap: wrap;
            gap: 15px;
        }
        h1 { color: #333; font-size: 32px; margin-bottom: 10px; }
        .subtitle { color: #666; font-size: 16px; }
        .content { padding: 30px; }
        .loading { text-align: center; padding: 40px; color: #666; font-size: 18px; }
        .no-bakes { text-align: center; padding: 60px; color: #666; }
        .search-filter {
            display: flex;
            gap: 10px;
            margin-bottom: 20px;
            flex-wrap: wrap;
        }
        .search-input {
            flex: 1;
            min-width: 200px;
            padding: 12px;
            border: 2px solid #e5e7eb;
            border-radius: 8px;
            font-size: 14px;
        }
        .filter-btn {
            padding: 12px 20px;
            border: 2px solid #e5e7eb;
            border-radius: 8px;
            background: white;
            cursor: pointer;
            font-size: 14px;
            transition: all 0.2s;
        }
        .filter-btn.active {
            background: #f093fb;
            color: white;
            border-color: #f093fb;
        }
        .bake-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
            gap: 20px;
        }
        .bake-card {
            background: #f9fafb;
            border-radius: 12px;
            padding: 20px;
            cursor: pointer;
            transition: all 0.2s;
            border: 2px solid transparent;
        }
        .bake-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(0,0,0,0.1);
            border-color: #f093fb;
        }
        .bake-date {
            font-size: 12px;
            color: #666;
            margin-bottom: 8px;
        }
        .bake-title {
            font-size: 18px;
            font-weight: 600;
            color: #333;
            margin-bottom: 12px;
        }
        .bake-stats {
            display: flex;
            gap: 15px;
            margin-bottom: 12px;
            flex-wrap: wrap;
        }
        .stat {
            font-size: 14px;
            color: #666;
        }
        .stat strong {
            color: #333;
        }
        .bake-status {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 12px;
            font-size: 12px;
            font-weight: 600;
        }
        .status-complete {
            background: #dcfce7;
            color: #166534;
        }
        .status-in-progress {
            background: #fef3c7;
            color: #854d0e;
        }
        .assessment {
            margin-top: 12px;
            padding-top: 12px;
            border-top: 1px solid #e5e7eb;
        }
        .score-badge {
            display: inline-block;
            background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
            color: white;
            padding: 8px 16px;
            border-radius: 20px;
            font-size: 18px;
            font-weight: 700;
        }
        .proof-badge {
            display: inline-block;
            padding: 4px 8px;
            border-radius: 8px;
            font-size: 12px;
            margin-left: 8px;
        }
        .proof-good { background: #dcfce7; color: #166534; }
        .proof-under { background: #fef3c7; color: #854d0e; }
        .proof-over { background: #fee2e2; color: #991b1b; }
        .stats-summary {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
            gap: 15px;
            margin-bottom: 30px;
            padding: 20px;
            background: #f9fafb;
            border-radius: 12px;
        }
        .summary-item {
            text-align: center;
        }
        .summary-value {
            font-size: 24px;
            font-weight: 700;
            color: #f093fb;
        }
        .summary-label {
            font-size: 12px;
            color: #666;
            margin-top: 4px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div>
                <h1>📚 Bake History</h1>
                <p class="subtitle" id="subtitle">Loading bakes...</p>
            </div>
        </div>
        <div class="content">
            <div class="loading" id="loading">Loading bake history...</div>
            <div id="history-content" style="display: none;">
                <div class="stats-summary" id="stats-summary"></div>
                <div class="search-filter">
                    <input type="text" class="search-input" id="searchInput" placeholder="Search bakes by date or notes...">
                    <button class="filter-btn active" onclick="filterBakes('all')">All</button>
                    <button class="filter-btn" onclick="filterBakes('complete')">Completed</button>
                    <button class="filter-btn" onclick="filterBakes('in-progress')">In Progress</button>
                </div>
                <div class="bake-grid" id="bakeGrid"></div>
            </div>
            <div id="no-bakes" class="no-bakes" style="display: none;">
                <h2>No Bakes Yet</h2>
                <p>Start your first bake to see history</p>
            </div>
        </div>
    </div>

    <script>
        let allBakes = [];
        let currentFilter = 'all';

        async function loadHistory() {
            try {
                const response = await fetch('/api/bakes');
                allBakes = await response.json();

                if (!allBakes || allBakes.length === 0) {
                    document.getElementById('loading').style.display = 'none';
                    document.getElementById('no-bakes').style.display = 'block';
                    return;
                }

                document.getElementById('loading').style.display = 'none';
                document.getElementById('history-content').style.display = 'block';
                document.getElementById('subtitle').textContent = 'Viewing ' + allBakes.length + ' bakes';

                displayStatsSummary();
                displayBakes(allBakes);

                // Add search listener
                document.getElementById('searchInput').addEventListener('input', handleSearch);
            } catch (error) {
                console.error('Error loading history:', error);
                document.getElementById('loading').innerHTML = 'Error loading bake history';
            }
        }

        function displayStatsSummary() {
            const completed = allBakes.filter(b => b.completed).length;
            const totalEvents = allBakes.reduce((sum, b) => sum + b.event_count, 0);
            const avgEvents = totalEvents / allBakes.length;

            const scores = allBakes
                .filter(b => b.assessment && b.assessment.score)
                .map(b => b.assessment.score);
            const avgScore = scores.length > 0
                ? (scores.reduce((a, b) => a + b, 0) / scores.length).toFixed(1)
                : 'N/A';

            const html =
                '<div class="summary-item"><div class="summary-value">' + allBakes.length + '</div><div class="summary-label">Total Bakes</div></div>' +
                '<div class="summary-item"><div class="summary-value">' + completed + '</div><div class="summary-label">Completed</div></div>' +
                '<div class="summary-item"><div class="summary-value">' + avgEvents.toFixed(0) + '</div><div class="summary-label">Avg Events</div></div>' +
                '<div class="summary-item"><div class="summary-value">' + avgScore + '</div><div class="summary-label">Avg Score</div></div>';

            document.getElementById('stats-summary').innerHTML = html;
        }

        function displayBakes(bakes) {
            const grid = document.getElementById('bakeGrid');
            grid.innerHTML = '';

            bakes.forEach(bake => {
                const card = document.createElement('div');
                card.className = 'bake-card';
                card.onclick = () => window.location.href = '/view/status?date=' + encodeURIComponent(bake.date);

                let html = '<div class="bake-date">' + bake.date + '</div>';
                html += '<div class="bake-title">' + bake.start_time + '</div>';
                html += '<div class="bake-stats">';
                html += '<span class="stat"><strong>' + bake.event_count + '</strong> events</span>';
                html += '</div>';

                if (bake.completed) {
                    html += '<span class="bake-status status-complete">✓ Completed</span>';
                    if (bake.end_time) {
                        html += '<div class="bake-date" style="margin-top: 8px;">Finished: ' + bake.end_time + '</div>';
                    }
                } else {
                    html += '<span class="bake-status status-in-progress">⏳ In Progress</span>';
                }

                if (bake.assessment) {
                    html += '<div class="assessment">';
                    if (bake.assessment.score) {
                        html += '<span class="score-badge">' + bake.assessment.score + '/10</span>';
                    }
                    if (bake.assessment.proof_level) {
                        const proofClass = bake.assessment.proof_level === 'good' ? 'proof-good' :
                                         bake.assessment.proof_level === 'underproofed' ? 'proof-under' : 'proof-over';
                        html += '<span class="proof-badge ' + proofClass + '">' + bake.assessment.proof_level + '</span>';
                    }
                    if (bake.assessment.notes) {
                        html += '<div class="bake-date" style="margin-top: 8px;">📝 ' + bake.assessment.notes + '</div>';
                    }
                    html += '</div>';
                }

                card.innerHTML = html;
                grid.appendChild(card);
            });
        }

        function filterBakes(filter) {
            currentFilter = filter;

            // Update button states
            document.querySelectorAll('.filter-btn').forEach(btn => {
                btn.classList.remove('active');
            });
            event.target.classList.add('active');

            applyFilters();
        }

        function handleSearch() {
            applyFilters();
        }

        function applyFilters() {
            const searchTerm = document.getElementById('searchInput').value.toLowerCase();

            let filtered = allBakes;

            // Apply status filter
            if (currentFilter === 'complete') {
                filtered = filtered.filter(b => b.completed);
            } else if (currentFilter === 'in-progress') {
                filtered = filtered.filter(b => !b.completed);
            }

            // Apply search filter
            if (searchTerm) {
                filtered = filtered.filter(b => {
                    return b.date.toLowerCase().includes(searchTerm) ||
                           b.start_time.toLowerCase().includes(searchTerm) ||
                           (b.assessment && b.assessment.notes && b.assessment.notes.toLowerCase().includes(searchTerm));
                });
            }

            displayBakes(filtered);
        }

        // Load history on page load
        loadHistory();
    </script>
</body>
</html>`

const statusViewPageHTML = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Bake Status</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.0/dist/chart.umd.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/chartjs-adapter-date-fns@3.0.0/dist/chartjs-adapter-date-fns.bundle.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/chartjs-plugin-zoom@2.0.1/dist/chartjs-plugin-zoom.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/hammerjs@2.0.8/hammer.min.js"></script>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            margin-bottom: 20px;
        }
        .header {
            padding: 30px;
            border-bottom: 2px solid #f3f4f6;
        }
        h1 { color: #333; font-size: 32px; margin-bottom: 10px; }
        .subtitle { color: #666; font-size: 16px; }
        .content { padding: 30px; }
        .loading { text-align: center; padding: 40px; color: #666; font-size: 18px; }
        .chart-container { position: relative; height: 400px; margin-bottom: 30px; }
        .timeline { margin-top: 30px; }
        .event-item {
            padding: 15px;
            margin-bottom: 10px;
            background: #f9fafb;
            border-radius: 10px;
            border-left: 4px solid #667eea;
        }
        .event-time { font-size: 12px; color: #666; margin-bottom: 5px; }
        .event-name { font-size: 16px; font-weight: 600; color: #333; }
        .event-details { font-size: 14px; color: #666; margin-top: 5px; }
        .event-note {
            background: #fef3c7;
            padding: 10px;
            border-radius: 5px;
            margin-top: 8px;
            font-style: italic;
        }
        .event-image {
            margin-top: 10px;
            cursor: pointer;
        }
        .event-image img {
            max-width: 200px;
            height: auto;
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            transition: transform 0.2s;
        }
        .event-image img:hover {
            transform: scale(1.05);
        }
        .modal {
            display: none;
            position: fixed;
            z-index: 1000;
            left: 0;
            top: 0;
            width: 100%;
            height: 100%;
            background: rgba(0,0,0,0.9);
            align-items: center;
            justify-content: center;
        }
        .modal.active {
            display: flex;
        }
        .modal-content {
            max-width: 90%;
            max-height: 90%;
            object-fit: contain;
        }
        .modal-close {
            position: absolute;
            top: 20px;
            right: 35px;
            color: #f1f1f1;
            font-size: 40px;
            font-weight: bold;
            cursor: pointer;
        }
        .modal-close:hover {
            color: #bbb;
        }
        .controls {
            margin-bottom: 20px;
            display: flex;
            gap: 10px;
            flex-wrap: wrap;
        }
        .btn {
            padding: 10px 20px;
            border: none;
            border-radius: 8px;
            background: #667eea;
            color: white;
            font-size: 14px;
            font-weight: 600;
            cursor: pointer;
            transition: background 0.2s;
        }
        .btn:hover { background: #5568d3; }
        .btn-secondary {
            background: #e5e7eb;
            color: #333;
        }
        .btn-secondary:hover { background: #d1d5db; }
        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
            margin-bottom: 30px;
        }
        .stat-card {
            background: #f9fafb;
            padding: 20px;
            border-radius: 10px;
            text-align: center;
        }
        .stat-value { font-size: 32px; font-weight: 700; color: #667eea; }
        .stat-label { font-size: 14px; color: #666; margin-top: 5px; }
        .no-data { text-align: center; padding: 60px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📊 Bake Status</h1>
            <p class="subtitle" id="subtitle">Loading current bake...</p>
        </div>
        <div class="content">
            <div class="loading" id="loading">Loading bake data...</div>
            <div id="bake-content" style="display: none;">
                <div class="stats" id="stats"></div>

                <h3 style="margin-top: 20px; margin-bottom: 10px;">Fermentation Temperatures</h3>
                <div class="chart-container">
                    <canvas id="fermentChart"></canvas>
                </div>

                <h3 style="margin-top: 30px; margin-bottom: 10px;">Baking Temperatures</h3>
                <div class="chart-container">
                    <canvas id="bakeChart"></canvas>
                </div>

                <div class="timeline" id="timeline"></div>
                <div style="margin-top: 40px; padding-top: 20px; border-top: 2px solid #e5e7eb; text-align: center;">
                    <button class="btn" style="background: #dc2626; border-color: #dc2626;" onclick="deleteBake()">🗑️ Delete This Bake</button>
                    <p style="color: #666; font-size: 12px; margin-top: 8px;">This will move the bake to the trash directory</p>
                </div>
            </div>
            <div id="no-data" class="no-data" style="display: none;">
                <h2>No Bakes Found</h2>
                <p>Start a new bake to see status here</p>
            </div>
        </div>
    </div>

    <!-- Image Modal -->
    <div id="imageModal" class="modal" onclick="closeModal()">
        <span class="modal-close">&times;</span>
        <img class="modal-content" id="modalImage">
    </div>

    <script>
        let fermentChart;
        let bakeChart;
        let bakeData;

        async function loadBake() {
            try {
                // Check if a date parameter is provided in the URL
                const urlParams = new URLSearchParams(window.location.search);
                const date = urlParams.get('date');
                const apiUrl = date ? '/api/bake/' + date : '/api/bake/current';

                const response = await fetch(apiUrl);
                bakeData = await response.json();

                if (!bakeData.events || bakeData.events.length === 0) {
                    document.getElementById('loading').style.display = 'none';
                    document.getElementById('no-data').style.display = 'block';
                    return;
                }

                displayBake(bakeData);
            } catch (error) {
                console.error('Error loading bake:', error);
                document.getElementById('loading').innerHTML = 'Error loading bake data';
            }
        }

        function displayBake(bake) {
            document.getElementById('loading').style.display = 'none';
            document.getElementById('bake-content').style.display = 'block';

            // Update subtitle
            const startTime = new Date(bake.events[0].timestamp);
            const isComplete = bake.events[bake.events.length - 1].event === 'loaf-complete';
            document.getElementById('subtitle').textContent =
                'Started ' + startTime.toLocaleString() + (isComplete ? ' (Completed)' : ' (In Progress)');

            // Display stats
            displayStats(bake);

            // Display chart
            displayChart(bake);

            // Display timeline
            displayTimeline(bake);
        }

        function displayStats(bake) {
            const stats = document.getElementById('stats');
            const events = bake.events;

            const temps = events.filter(e => e.temp_f || e.dough_temp_f);
            const avgTemp = temps.length > 0
                ? temps.reduce((sum, e) => sum + (e.temp_f || e.dough_temp_f || 0), 0) / temps.length
                : 0;

            const duration = (new Date(events[events.length - 1].timestamp) - new Date(events[0].timestamp)) / (1000 * 60 * 60);
            const folds = events.filter(e => e.event === 'fold').length;
            const isComplete = events[events.length - 1].event === 'loaf-complete';

            let statsHTML = '<div class="stat-card"><div class="stat-value">' + events.length + '</div><div class="stat-label">Events</div></div>';
            statsHTML += '<div class="stat-card"><div class="stat-value">' + duration.toFixed(1) + 'h</div><div class="stat-label">Duration</div></div>';
            statsHTML += '<div class="stat-card"><div class="stat-value">' + folds + '</div><div class="stat-label">Folds</div></div>';
            if (avgTemp > 0) {
                statsHTML += '<div class="stat-card"><div class="stat-value">' + avgTemp.toFixed(1) + '°F</div><div class="stat-label">Avg Temp</div></div>';
            }
            if (isComplete && bake.assessment && bake.assessment.score) {
                statsHTML += '<div class="stat-card"><div class="stat-value">' + bake.assessment.score + '/10</div><div class="stat-label">Score</div></div>';
            }

            stats.innerHTML = statsHTML;
        }

        function displayChart(bake) {
            const fermentCtx = document.getElementById('fermentChart');
            const bakeCtx = document.getElementById('bakeChart');

            // Find oven-in event to split kitchen vs oven temps, dough vs loaf temps
            const ovenInIdx = bake.events.findIndex(e => e.event === 'oven-in');

            // Prepare data points
            const labels = [];
            const kitchenTemps = [];
            const doughTemps = [];
            const loafTemps = [];
            const ovenTemps = [];
            const notePoints = [];
            const eventAnnotations = [];

            bake.events.forEach((event, idx) => {
                const time = new Date(event.timestamp);
                labels.push(time);

                // Temperature data - split kitchen vs oven temps
                if (event.temp_f) {
                    if (ovenInIdx >= 0 && idx >= ovenInIdx) {
                        // After oven-in, temp_f is oven temp
                        ovenTemps.push(event.temp_f);
                        kitchenTemps.push(null);
                    } else {
                        // Before oven-in, temp_f is kitchen temp
                        kitchenTemps.push(event.temp_f);
                        ovenTemps.push(null);
                    }
                } else {
                    kitchenTemps.push(null);
                    ovenTemps.push(null);
                }

                // Separate dough temps (before oven) from loaf temps (during baking)
                if (event.dough_temp_f) {
                    if (ovenInIdx >= 0 && idx >= ovenInIdx) {
                        // After oven-in, dough_temp_f is loaf internal temp
                        loafTemps.push(event.dough_temp_f);
                        doughTemps.push(null);
                    } else {
                        // Before oven-in, dough_temp_f is dough temp
                        doughTemps.push(event.dough_temp_f);
                        loafTemps.push(null);
                    }
                } else {
                    doughTemps.push(null);
                    loafTemps.push(null);
                }

                // Note markers
                if (event.note) {
                    notePoints.push({
                        x: time,
                        y: event.temp_f || event.dough_temp_f || 70,
                        note: event.note
                    });
                }

                // Event markers
                if (['oven-in', 'oven-out', 'shaped', 'mixed', 'fridge-in'].includes(event.event)) {
                    eventAnnotations.push({
                        type: 'line',
                        xMin: time,
                        xMax: time,
                        borderColor: event.event === 'oven-in' ? 'rgba(220, 38, 38, 0.5)' : 'rgba(59, 130, 246, 0.5)',
                        borderWidth: 2,
                        borderDash: [5, 5],
                        label: {
                            display: true,
                            content: event.event,
                            position: 'start'
                        }
                    });
                }
            });

            if (fermentChart) fermentChart.destroy();
            if (bakeChart) bakeChart.destroy();

            // Fermentation Chart (Kitchen + Dough temps before oven-in)
            const fermentLabels = [];
            const fermentKitchen = [];
            const fermentDough = [];
            const fermentNotes = [];

            // Baking Chart (Oven + Loaf temps from oven-in onwards)
            const bakeLabels = [];
            const bakeOven = [];
            const bakeLoaf = [];
            const bakeNotes = [];

            // Stage events that should have labels
            const stageEvents = ['starter-out', 'fed', 'levain-ready', 'mixed', 'fold', 'shaped',
                                 'fridge-in', 'fridge-out', 'oven-in', 'remove-lid', 'oven-out'];

            bake.events.forEach((event, idx) => {
                const time = new Date(event.timestamp);
                const isStage = stageEvents.includes(event.event);

                if (ovenInIdx < 0 || idx < ovenInIdx) {
                    // Before oven-in: fermentation phase
                    fermentLabels.push(time);
                    // Store event name with temp for stage events
                    if (event.temp_f) {
                        fermentKitchen.push({
                            x: time,
                            y: event.temp_f,
                            stage: isStage ? event.event : null
                        });
                    } else {
                        fermentKitchen.push(null);
                    }
                    if (event.dough_temp_f) {
                        fermentDough.push({
                            x: time,
                            y: event.dough_temp_f,
                            stage: isStage ? event.event : null
                        });
                    } else {
                        fermentDough.push(null);
                    }
                    if (event.note) {
                        fermentNotes.push({
                            x: time,
                            y: 55,
                            note: event.note
                        });
                    }
                } else {
                    // From oven-in onwards: baking phase
                    bakeLabels.push(time);
                    // Store event name with temp for stage events
                    if (event.temp_f) {
                        bakeOven.push({
                            x: time,
                            y: event.temp_f,
                            stage: isStage ? event.event : null
                        });
                    } else {
                        bakeOven.push(null);
                    }
                    if (event.dough_temp_f) {
                        bakeLoaf.push({
                            x: time,
                            y: event.dough_temp_f,
                            stage: isStage ? event.event : null
                        });
                    } else {
                        bakeLoaf.push(null);
                    }
                    if (event.note) {
                        bakeNotes.push({
                            x: time,
                            y: 55,
                            note: event.note
                        });
                    }
                }
            });

            // Create Fermentation Chart
            fermentChart = new Chart(fermentCtx, {
                type: 'line',
                data: {
                    datasets: [
                        {
                            label: 'Kitchen Temp (°F)',
                            data: fermentKitchen,
                            borderColor: 'rgb(59, 130, 246)',
                            backgroundColor: 'rgba(59, 130, 246, 0.1)',
                            tension: 0.4,
                            spanGaps: true,
                            parsing: false
                        },
                        {
                            label: 'Dough Temp (°F)',
                            data: fermentDough,
                            borderColor: 'rgb(220, 38, 38)',
                            backgroundColor: 'rgba(220, 38, 38, 0.1)',
                            tension: 0.4,
                            spanGaps: true,
                            parsing: false
                        },
                        {
                            label: 'Notes',
                            data: fermentNotes,
                            type: 'scatter',
                            backgroundColor: 'rgb(251, 191, 36)',
                            pointRadius: 6,
                            pointHoverRadius: 8,
                            showLine: false
                        }
                    ]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    interaction: { mode: 'index', intersect: false },
                    scales: {
                        x: {
                            type: 'time',
                            time: { displayFormats: { minute: 'h:mm a', hour: 'MMM d ha', day: 'MMM d' } },
                            title: { display: true, text: 'Time' }
                        },
                        y: {
                            title: { display: true, text: 'Temperature (°F)' },
                            min: 50,
                            max: 85
                        }
                    },
                    plugins: {
                        tooltip: {
                            callbacks: {
                                afterLabel: function(context) {
                                    if (context.dataset.label === 'Notes' && context.raw.note) {
                                        return context.raw.note;
                                    }
                                    if (context.raw && context.raw.stage) {
                                        return 'Event: ' + context.raw.stage;
                                    }
                                    return '';
                                }
                            }
                        }
                    }
                }
            });

            // Create Baking Chart
            bakeChart = new Chart(bakeCtx, {
                type: 'line',
                data: {
                    datasets: [
                        {
                            label: 'Oven Temp (°F)',
                            data: bakeOven,
                            borderColor: 'rgb(249, 115, 22)',
                            backgroundColor: 'rgba(249, 115, 22, 0.1)',
                            tension: 0.4,
                            spanGaps: true,
                            parsing: false
                        },
                        {
                            label: 'Loaf Internal Temp (°F)',
                            data: bakeLoaf,
                            borderColor: 'rgb(147, 51, 234)',
                            backgroundColor: 'rgba(147, 51, 234, 0.1)',
                            tension: 0.4,
                            spanGaps: true,
                            parsing: false
                        },
                        {
                            label: 'Notes',
                            data: bakeNotes,
                            type: 'scatter',
                            backgroundColor: 'rgb(251, 191, 36)',
                            pointRadius: 6,
                            pointHoverRadius: 8,
                            showLine: false
                        }
                    ]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    interaction: { mode: 'index', intersect: false },
                    scales: {
                        x: {
                            type: 'time',
                            time: { displayFormats: { minute: 'h:mm a', hour: 'MMM d ha', day: 'MMM d' } },
                            title: { display: true, text: 'Time' }
                        },
                        y: {
                            title: { display: true, text: 'Temperature (°F)' },
                            min: 50,
                            max: 550
                        }
                    },
                    plugins: {
                        tooltip: {
                            callbacks: {
                                afterLabel: function(context) {
                                    if (context.dataset.label === 'Notes' && context.raw.note) {
                                        return context.raw.note;
                                    }
                                    if (context.raw && context.raw.stage) {
                                        return 'Event: ' + context.raw.stage;
                                    }
                                    return '';
                                }
                            }
                        }
                    }
                }
            });
        }

        function displayTimeline(bake) {
            const timeline = document.getElementById('timeline');
            let html = '<h2 style="margin-bottom: 20px;">Event Timeline</h2>';

            bake.events.forEach(event => {
                const time = new Date(event.timestamp);
                html += '<div class="event-item">';
                html += '<div class="event-time">' + time.toLocaleString() + '</div>';
                html += '<div class="event-name">' + event.event + '</div>';

                const details = [];
                if (event.temp_f) details.push('Kitchen: ' + event.temp_f + '°F');
                if (event.dough_temp_f) details.push('Dough: ' + event.dough_temp_f + '°F');
                if (event.fold_count) details.push('Fold #' + event.fold_count);

                if (details.length > 0) {
                    html += '<div class="event-details">' + details.join(' • ') + '</div>';
                }

                if (event.note) {
                    html += '<div class="event-note">📝 ' + event.note + '</div>';
                }

                // Display image thumbnail if present
                if (event.image) {
                    const imageUrl = '/images/' + bake.filename + '/' + event.image;
                    html += '<div class="event-image" onclick="openModal(\'' + imageUrl + '\')">';
                    html += '<img src="' + imageUrl + '" alt="Event photo" title="Click to enlarge">';
                    html += '</div>';
                }

                html += '</div>';
            });

            timeline.innerHTML = html;
        }

        function openModal(imageUrl) {
            const modal = document.getElementById('imageModal');
            const modalImg = document.getElementById('modalImage');
            modalImg.src = imageUrl;
            modal.classList.add('active');
        }

        function closeModal() {
            const modal = document.getElementById('imageModal');
            modal.classList.remove('active');
        }

        // Close modal with Escape key
        document.addEventListener('keydown', function(event) {
            if (event.key === 'Escape') {
                closeModal();
            }
        });

        function resetZoom() {
            if (chart) {
                chart.resetZoom();
                // Restore original Y-axis settings
                chart.options.scales.y.min = 60;
                chart.options.scales.y.max = 500;
                chart.update();
            }
        }

        function zoomToBaking() {
            if (!bakeData) return;

            const ovenInEvent = bakeData.events.find(e => e.event === 'oven-in');
            const ovenOutEvent = bakeData.events.find(e => e.event === 'oven-out');

            if (ovenInEvent) {
                const start = new Date(ovenInEvent.timestamp);
                const end = ovenOutEvent ? new Date(ovenOutEvent.timestamp) : new Date(start.getTime() + 60*60*1000);

                // Find all temps during baking period
                const bakingEvents = bakeData.events.filter(e => {
                    const t = new Date(e.timestamp);
                    return t >= start && t <= end;
                });

                const temps = [];
                bakingEvents.forEach(e => {
                    if (e.temp_f) temps.push(e.temp_f);
                    if (e.dough_temp_f) temps.push(e.dough_temp_f);
                });

                if (temps.length === 0) return;

                // Calculate Y-axis range with padding
                let minTemp = Math.min(...temps);
                let maxTemp = Math.max(...temps);

                // Add 10% padding or at least 20 degrees
                const range = maxTemp - minTemp;
                const padding = Math.max(range * 0.1, 20);

                minTemp = Math.floor(minTemp - padding);
                maxTemp = Math.ceil(maxTemp + padding);

                // Zoom both axes
                chart.zoomScale('x', {min: start, max: end}, 'default');
                chart.options.scales.y.min = minTemp;
                chart.options.scales.y.max = maxTemp;
                chart.update();
            }
        }

        async function deleteBake() {
            // Get the date parameter if viewing a specific bake
            const urlParams = new URLSearchParams(window.location.search);
            const date = urlParams.get('date');

            if (!date) {
                alert('Cannot delete the current in-progress bake. Only completed bakes can be deleted.');
                return;
            }

            // Confirm deletion
            const confirmMsg = 'Are you sure you want to delete this bake?\n\n' +
                              'Date: ' + date + '\n\n' +
                              'This will move it to the trash directory.';

            if (!confirm(confirmMsg)) {
                return;
            }

            try {
                const response = await fetch('/api/bake/' + encodeURIComponent(date), {
                    method: 'DELETE'
                });

                const result = await response.json();

                if (response.ok) {
                    alert('Bake deleted successfully and moved to trash.');
                    window.location.href = '/view/history';
                } else {
                    alert('Error deleting bake: ' + (result.error || 'Unknown error'));
                }
            } catch (error) {
                console.error('Error deleting bake:', error);
                alert('Error deleting bake: ' + error.message);
            }
        }

        // Load bake on page load
        loadBake();
    </script>
</body>
</html>`
