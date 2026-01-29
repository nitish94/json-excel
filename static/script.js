document.addEventListener('DOMContentLoaded', () => {
    const container = document.getElementById('spreadsheet-container');
    const saveBtn = document.getElementById('saveBtn');
    const refreshBtn = document.getElementById('refreshBtn');
    const uploadBtn = document.getElementById('uploadBtn');
    const downloadBtn = document.getElementById('downloadBtn');
    const fileUpload = document.getElementById('fileUpload');
    const statusEl = document.getElementById('status');

    let currentData = [];

    // Validations
    const MAX_KPIS = 10;

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

            // Refund/Reload data
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
        window.location.href = '/api/download';
    });

    async function fetchData() {
        showStatus('Loading...', 'normal');
        try {
            const response = await fetch('/api/data');
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

            const response = await fetch('/api/data', {
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

    function renderTable(data) {
        container.innerHTML = '';
        if (!Array.isArray(data) || data.length === 0) {
            container.innerHTML = '<p style="text-align:center; opacity:0.6;">No data found or empty.</p>';
            return;
        }

        const table = document.createElement('table');
        const thead = document.createElement('thead');
        const tbody = document.createElement('tbody');

        // Extract all unique keys for columns (Root headers)
        let allKeys = new Set();
        data.forEach(obj => Object.keys(obj).forEach(k => allKeys.add(k)));
        let headers = Array.from(allKeys);

        // Enforce MAX 10 KPIs constraint visual warning or slicing? 
        // The backend enforces it on save. Here we just render what we have.
        // But if we want to "automatically become null" for missing, we handle that in rows.

        // Header Row
        const trHead = document.createElement('tr');
        headers.forEach(key => {
            const th = document.createElement('th');
            th.textContent = key;
            trHead.appendChild(th);
        });
        thead.appendChild(trHead);

        // Data Rows
        data.forEach((rowObj, rowIndex) => {
            const tr = document.createElement('tr');
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
                    // Object treated as one-row nested table or just JSON string?
                    // Use case says "only one nested json is allowed".
                    // Let's assume nested object is also rendered as a small table or flat?
                    // Let's treat it same as array of objects of size 1 for uniformity? 
                    // Or just render text for now if simple object.
                    // Let's recursively render simple object as table too.
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
                        // Try to convert to number if it looks like one?
                        // Or keep strict types?
                        // For a generic editor, string is safest. 
                        // But if user wants Excel-like, numbers should be numbers.
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
