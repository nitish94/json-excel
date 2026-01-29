document.addEventListener('DOMContentLoaded', () => {
    const container = document.getElementById('spreadsheet-container');
    const saveBtn = document.getElementById('saveBtn');
    const refreshBtn = document.getElementById('refreshBtn');
    const uploadBtn = document.getElementById('uploadBtn');
    const downloadBtn = document.getElementById('downloadBtn');
    const addRowBtn = document.getElementById('addRowBtn');
    const addColumnBtn = document.getElementById('addColumnBtn');
    const fileUpload = document.getElementById('fileUpload');
    const statusEl = document.getElementById('status');

    let currentData = [];

    // Validations
    const MAX_KPIS = 10;

    // Get ID from URL or default to 'demo'
    const urlParams = new URLSearchParams(window.location.search);
    let currentId = urlParams.get('id') || 'demo';

    // Update URL if not present (optional, but good for sharing)
    if (!urlParams.has('id')) {
        const newUrl = window.location.protocol + "//" + window.location.host + window.location.pathname + '?id=' + currentId;
        window.history.replaceState({ path: newUrl }, '', newUrl);
    }

    uploadBtn.addEventListener('click', () => {
        fileUpload.click();
    });

    displayUploadFileName = () => {
        // Optional: Update status when file picked
    }

    fileUpload.addEventListener('change', async (e) => {
        if (fileUpload.files.length === 0) return;

        const file = fileUpload.files[0];
        const formData = new FormData();
        formData.append('file', file);

        showStatus('Uploading...', 'normal');

        try {
            const response = await fetch('/api/upload', {
                method: 'POST',
                body: formData
            });

            if (!response.ok) {
                const errText = await response.text();
                throw new Error(errText);
            }

            const result = await response.json();
            showStatus(result.message || 'Uploaded successfully', 'success');

            // Update ID and Reload
            if (result.id) {
                currentId = result.id;
                // Update URL
                const newUrl = window.location.protocol + "//" + window.location.host + window.location.pathname + '?id=' + currentId;
                window.history.pushState({ path: newUrl }, '', newUrl);
            }

            await fetchData();
            // Clear input
            fileUpload.value = '';

        } catch (error) {
            console.error(error);
            showStatus(`Upload failed: ${error.message}`, 'error');
            fileUpload.value = '';
        }
    });

    downloadBtn.addEventListener('click', () => {
        window.location.href = `/api/download?id=${currentId}`;
    });

    async function fetchData() {
        showStatus('Loading...', 'normal');
        try {
            const response = await fetch(`/api/data?id=${currentId}`);
            if (!response.ok) throw new Error('Failed to fetch data');
            currentData = await response.json();
            renderTable(currentData);
            showStatus('Loaded', 'success');
        } catch (error) {
            console.error(error);
            showStatus('Error loading data', 'error');
        }
    }

    async function saveData() {
        showStatus('Saving...', 'normal');
        try {
            // Reconstruct JSON from DOM or use currentData if we bound it properly.
            // For a robust implementation, it's better to update currentData on input change.
            console.log("Saving data:", currentData);

            const response = await fetch(`/api/data?id=${currentId}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(currentData)
            });

            if (!response.ok) {
                const errText = await response.text();
                throw new Error(errText);
            }

            showStatus('Saved successfully!', 'success');
        } catch (error) {
            console.error(error);
            showStatus(`Save failed: ${error.message}`, 'error');
        }
    }

    function showStatus(msg, type) {
        statusEl.textContent = msg;
        if (type === 'error') statusEl.style.color = '#ff6b6b';
        else if (type === 'success') statusEl.style.color = '#4ade80';
        else statusEl.style.color = '#94a3b8';

        setTimeout(() => {
            if (statusEl.textContent === msg) statusEl.textContent = '';
        }, 5000);
    }


    // ... (existing code)

    addRowBtn.addEventListener('click', () => {
        if (!currentData || currentData.length === 0) {
            // Initialize with one empty object if empty
            currentData = [{}];
        } else {
            // Create a new object with keys from the first row (or all keys)
            const firstRow = currentData[0];
            const newRow = {};
            Object.keys(firstRow).forEach(key => newRow[key] = "");
            currentData.push(newRow);
        }
        renderTable(currentData);
        saveData(); // Auto-save on structural change? Or let user save? User request implies "Editor", usually auto-save or explicit. we have explicit save button. Let's start with just UI update.
    });

    addColumnBtn.addEventListener('click', () => {
        const newKey = prompt("Enter new column name:");
        if (!newKey) return;

        // Validation: Max 10 keys
        // We need to check unlimited keys or per object? 
        // Logic says "Max 10 keys per object".
        // Let's check max keys of first object + 1
        if (currentData.length > 0) {
            const currentKeys = Object.keys(currentData[0]);
            if (currentKeys.length >= 10) {
                alert("Maximum 10 columns allowed.");
                return;
            }
            if (currentKeys.includes(newKey)) {
                alert("Column already exists.");
                return;
            }
        }

        // Add key to all rows
        if (currentData.length === 0) {
            currentData = [{ [newKey]: "" }];
        } else {
            currentData.forEach(row => {
                row[newKey] = "";
            });
        }
        renderTable(currentData);
    });

    // ... (existing listeners)

    function deleteRow(index) {
        if (confirm("Delete this row?")) {
            currentData.splice(index, 1);
            renderTable(currentData);
        }
    }

    function deleteColumn(key) {
        if (confirm(`Delete column "$\{key}"?`)) {
            currentData.forEach(row => {
                delete row[key];
            });
            renderTable(currentData);
        }
    }

    function renderTable(data) {
        container.innerHTML = '';
        if (!Array.isArray(data) || data.length === 0) {
            container.innerHTML = '<p style="text-align:center; opacity:0.6;">No data found. Click "+ Row" to start.</p>';
            return;
        }

        const table = document.createElement('table');
        const thead = document.createElement('thead');
        const tbody = document.createElement('tbody');

        // Extract all unique keys for columns (Root headers)
        let allKeys = new Set();
        data.forEach(obj => Object.keys(obj).forEach(k => allKeys.add(k)));
        let headers = Array.from(allKeys);

        // Header Row
        const trHead = document.createElement('tr');

        // Action Header (Empty)
        const thAction = document.createElement('th');
        thAction.style.width = '50px';
        trHead.appendChild(thAction);

        headers.forEach(key => {
            const th = document.createElement('th');
            // Container for text and delete btn
            const div = document.createElement('div');
            div.style.display = 'flex';
            div.style.justifyContent = 'space-between';
            div.style.alignItems = 'center';

            const span = document.createElement('span');
            span.textContent = key;

            const delBtn = document.createElement('button');
            delBtn.className = 'delete-col-btn';
            delBtn.innerHTML = '&times;';
            delBtn.title = 'Delete Column';
            delBtn.onclick = (e) => { e.stopPropagation(); deleteColumn(key); };

            div.appendChild(span);
            div.appendChild(delBtn);
            th.appendChild(div);
            trHead.appendChild(th);
        });
        thead.appendChild(trHead);

        // Data Rows
        data.forEach((rowObj, rowIndex) => {
            const tr = document.createElement('tr');

            // Delete Row Cell
            const tdAction = document.createElement('td');
            const delRowBtn = document.createElement('button');
            delRowBtn.className = 'delete-row-btn';
            delRowBtn.innerHTML = '&times;';
            delRowBtn.title = 'Delete Row';
            delRowBtn.onclick = () => deleteRow(rowIndex);
            tdAction.appendChild(delRowBtn);
            tr.appendChild(tdAction);

            headers.forEach(key => {
                const td = document.createElement('td');
                const value = rowObj[key];

                // Check if value is a "Nested Table" (Array of Objects)
                if (Array.isArray(value) && value.length > 0 && typeof value[0] === 'object') {
                    // Render Nested Table
                    const nestedDiv = document.createElement('div');
                    nestedDiv.className = 'nested-table-container';
                    nestedDiv.appendChild(createNestedTable(value, rowIndex, key));
                    td.appendChild(nestedDiv);
                } else if (typeof value === 'object' && value !== null && !Array.isArray(value)) {
                    // Object treated as one-row nested table
                    const nestedDiv = document.createElement('div');
                    nestedDiv.className = 'nested-table-container';
                    nestedDiv.appendChild(createNestedTable([value], rowIndex, key, true)); // Wrap in array
                    td.appendChild(nestedDiv);
                } else {
                    // Primitive Value -> Input
                    const input = document.createElement('input');
                    input.className = 'cell-input';
                    input.type = 'text';
                    input.value = value === undefined || value === null ? '' : value;

                    input.addEventListener('input', (e) => {
                        let val = e.target.value;
                        if (!isNaN(val) && val.trim() !== '') {
                            val = Number(val);
                        }
                        currentData[rowIndex][key] = val;
                    });

                    td.appendChild(input);
                }
                tr.appendChild(td);
            });
            tbody.appendChild(tr);
        });

        table.appendChild(thead);
        table.appendChild(tbody);
        container.appendChild(table);
    }

    function createNestedTable(nestedData, parentIndex, parentKey, isSingleObject = false) {
        const table = document.createElement('table');
        table.className = 'nested-table';
        const thead = document.createElement('thead');
        const tbody = document.createElement('tbody');

        let nestedKeys = new Set();
        nestedData.forEach(obj => Object.keys(obj).forEach(k => nestedKeys.add(k)));
        let headers = Array.from(nestedKeys);

        // Header
        const trHead = document.createElement('tr');
        headers.forEach(key => {
            const th = document.createElement('th');
            th.textContent = key;
            trHead.appendChild(th);
        });
        thead.appendChild(trHead);

        // Rows
        nestedData.forEach((rowObj, nestedIndex) => {
            const tr = document.createElement('tr');
            headers.forEach(key => {
                const td = document.createElement('td');
                const value = rowObj[key];

                const input = document.createElement('input');
                input.className = 'cell-input';
                input.value = value === undefined || value === null ? '' : value;

                input.addEventListener('input', (e) => {
                    let val = e.target.value;
                    if (!isNaN(val) && val.trim() !== '') {
                        val = Number(val);
                    }

                    if (isSingleObject) {
                        // It was a single object, so we are editing the object itself at currentData[parentIndex][parentKey]
                        currentData[parentIndex][parentKey][key] = val;
                    } else {
                        // It is an array
                        currentData[parentIndex][parentKey][nestedIndex][key] = val;
                    }
                });

                td.appendChild(input);
                tr.appendChild(td);
            });
            tbody.appendChild(tr);
        });

        table.appendChild(thead);
        table.appendChild(tbody);
        return table;
    }

    saveBtn.addEventListener('click', saveData);
    refreshBtn.addEventListener('click', fetchData);

    // Initial Load
    fetchData();
});
