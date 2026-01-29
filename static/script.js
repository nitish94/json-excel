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
    const MAX_KPIS = 20;

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
            // Create a new object with keys from the first row, copying structure
            const firstRow = currentData[0];
            const newRow = {};
            Object.keys(firstRow).forEach(key => {
                const val = firstRow[key];
                if (Array.isArray(val) && val.length > 0 && typeof val[0] === 'object') {
                    // Nested array of objects: create one empty object with same keys
                    const firstObj = val[0];
                    const emptyObj = {};
                    Object.keys(firstObj).forEach(k => emptyObj[k] = "");
                    newRow[key] = [emptyObj];
                } else if (Array.isArray(val)) {
                    newRow[key] = [];
                } else if (typeof val === 'object' && val !== null) {
                    // Single object
                    const emptyObj = {};
                    Object.keys(val).forEach(k => emptyObj[k] = "");
                    newRow[key] = emptyObj;
                } else {
                    newRow[key] = "";
                }
            });
            currentData.push(newRow);
        }
        renderTable(currentData);
    });

    addColumnBtn.addEventListener('click', () => {
        const newKey = prompt("Enter new column name:");
        if (!newKey) return;

        let isNested = false;
        let subKeys = [];
        const type = prompt("Enter column type: 'primitive' (default) or 'nested':");
        if (type && type.toLowerCase() === 'nested') {
            isNested = true;
            const subKeysStr = prompt("Enter sub-column names separated by comma (e.g. metric,value):");
            if (!subKeysStr) return;
            subKeys = subKeysStr.split(',').map(s => s.trim()).filter(s => s);
            if (subKeys.length === 0) return;
        }

        // Validation: Max keys per object
        if (currentData.length > 0) {
            const currentKeys = Object.keys(currentData[0]);
            if (currentKeys.length >= MAX_KPIS) {
                alert(`Maximum ${MAX_KPIS} columns allowed.`);
                return;
            }
            if (currentKeys.includes(newKey)) {
                alert("Column already exists.");
                return;
            }
        }

        let newValue;
        if (isNested) {
            const emptyObj = {};
            subKeys.forEach(k => emptyObj[k] = "");
            newValue = [emptyObj];
        } else {
            newValue = "";
        }

        // Add key to all rows
        if (currentData.length === 0) {
            currentData = [{ [newKey]: newValue }];
        } else {
            currentData.forEach(row => {
                row[newKey] = newValue;
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
        if (confirm(`Delete column "${key}"?`)) {
            currentData.forEach(row => {
                delete row[key];
            });
            renderTable(currentData);
        }
    }

    function addNestedRow(parentIndex, parentKey, isSingleObject, headers) {
        if (isSingleObject) {
            // For single object, perhaps not add row, or convert to array
            // For simplicity, if single object, maybe alert not supported
            alert("Cannot add rows to single object nested.");
            return;
        }
        // Add a new empty object to the array
        const emptyObj = {};
        headers.forEach(key => emptyObj[key] = "");
        currentData[parentIndex][parentKey].push(emptyObj);
        renderTable(currentData);
    }

    function deleteNestedRow(parentIndex, parentKey, nestedIndex, isSingleObject) {
        if (isSingleObject) {
            alert("Cannot delete single object nested.");
            return;
        }
        if (confirm("Delete this nested row?")) {
            currentData[parentIndex][parentKey].splice(nestedIndex, 1);
            renderTable(currentData);
        }
    }

    function renderTable(data) {
        container.innerHTML = '';
        if (!Array.isArray(data) || data.length === 0) {
            const emptyDiv = document.createElement('div');
            emptyDiv.style.textAlign = 'center';
            emptyDiv.style.opacity = '0.6';
            emptyDiv.innerHTML = '<p>No data found. Start by adding columns or rows.</p>';
            const addColBtn = document.createElement('button');
            addColBtn.textContent = '+ Add Column';
            addColBtn.onclick = () => addColumnBtn.click();
            emptyDiv.appendChild(addColBtn);
            container.appendChild(emptyDiv);
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
        const container = document.createElement('div');
        container.className = 'nested-table-wrapper';

        // Add Row Button for nested
        const addNestedRowBtn = document.createElement('button');
        addNestedRowBtn.className = 'add-nested-row-btn';
        addNestedRowBtn.textContent = '+ Add Nested Row';
        addNestedRowBtn.onclick = () => addNestedRow(parentIndex, parentKey, isSingleObject, headers);
        container.appendChild(addNestedRowBtn);

        const table = document.createElement('table');
        table.className = 'nested-table';
        const thead = document.createElement('thead');
        const tbody = document.createElement('tbody');

        let nestedKeys = new Set();
        nestedData.forEach(obj => Object.keys(obj).forEach(k => nestedKeys.add(k)));
        let headers = Array.from(nestedKeys);

        // Header
        const trHead = document.createElement('tr');
        // Action column for delete
        const thAction = document.createElement('th');
        thAction.style.width = '30px';
        trHead.appendChild(thAction);
        headers.forEach(key => {
            const th = document.createElement('th');
            th.textContent = key;
            trHead.appendChild(th);
        });
        thead.appendChild(trHead);

        // Rows
        nestedData.forEach((rowObj, nestedIndex) => {
            const tr = document.createElement('tr');
            // Delete button
            const tdAction = document.createElement('td');
            const delBtn = document.createElement('button');
            delBtn.className = 'delete-nested-row-btn';
            delBtn.innerHTML = '&times;';
            delBtn.title = 'Delete Nested Row';
            delBtn.onclick = () => deleteNestedRow(parentIndex, parentKey, nestedIndex, isSingleObject);
            tdAction.appendChild(delBtn);
            tr.appendChild(tdAction);
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
        container.appendChild(table);
        return container;
    }

    saveBtn.addEventListener('click', saveData);
    refreshBtn.addEventListener('click', fetchData);

    // Initial Load
    fetchData();
});
