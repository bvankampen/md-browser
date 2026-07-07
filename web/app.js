const treeContainer = document.getElementById('tree-container');
const toggleAllCheckbox = document.getElementById('toggle-all');
const toggleAutofreshCheckbox = document.getElementById('toggle-autofresh');
const placeholderView = document.getElementById('placeholder-view');
const markdownView = document.getElementById('markdown-view');
const filePathDisplay = document.getElementById('file-path-display');
const markdownContent = document.getElementById('markdown-content');
const btnCopy = document.getElementById('btn-copy');
const btnExpandAll = document.getElementById('btn-expand-all');
const btnCollapseAll = document.getElementById('btn-collapse-all');
const btnRefresh = document.getElementById('btn-refresh');
const btnToggleSource = document.getElementById('btn-toggle-source');
const searchInput = document.getElementById('search-input');
const btnSearchClear = document.getElementById('btn-search-clear');
const searchResultsContainer = document.getElementById('search-results-container');

let currentActivePath = '';
let currentFileData = null;
let showingSource = false;
const expandedPaths = new Set();
let autofreshIntervalId = null;
let refreshIntervalSec = 5;

// Save show all state in localStorage
const storedShowAll = localStorage.getItem('md-browser-show-all');
if (storedShowAll !== null) {
    toggleAllCheckbox.checked = storedShowAll === 'true';
}

// Save auto-refresh state in localStorage (defaults to true for great UX!)
const storedAutofresh = localStorage.getItem('md-browser-auto-refresh');
if (storedAutofresh !== null) {
    toggleAutofreshCheckbox.checked = storedAutofresh === 'true';
} else {
    toggleAutofreshCheckbox.checked = true; // Enabled by default!
}

toggleAllCheckbox.addEventListener('change', () => {
    localStorage.setItem('md-browser-show-all', toggleAllCheckbox.checked);
    refreshTree();
});

toggleAutofreshCheckbox.addEventListener('change', () => {
    localStorage.setItem('md-browser-auto-refresh', toggleAutofreshCheckbox.checked);
    setupAutofresh();
});

btnCopy.addEventListener('click', () => {
    if (currentActivePath) {
        navigator.clipboard.writeText(currentActivePath).then(() => {
            const btnText = btnCopy.querySelector('span');
            const origText = btnText.innerText;
            btnText.innerText = 'Copied!';
            setTimeout(() => {
                btnText.innerText = origText;
            }, 1500);
        });
    }
});

btnExpandAll.addEventListener('click', (e) => {
    e.preventDefault();
    expandAllNode(treeContainer);
});

btnCollapseAll.addEventListener('click', (e) => {
    e.preventDefault();
    collapseAll();
});

btnRefresh.addEventListener('click', (e) => {
    e.preventDefault();
    refreshTree();
});

btnToggleSource.addEventListener('click', () => {
    if (!currentFileData || !currentFileData.is_markdown) return;
    showingSource = !showingSource;
    updateMarkdownDisplay();
});

let searchTimeoutId = null;
searchInput.addEventListener('input', () => {
    if (searchTimeoutId) {
        clearTimeout(searchTimeoutId);
    }

    const query = searchInput.value.trim();
    if (query === '') {
        btnSearchClear.classList.add('hidden');
        searchResultsContainer.classList.add('hidden');
        treeContainer.classList.remove('hidden');
        return;
    }

    btnSearchClear.classList.remove('hidden');
    searchTimeoutId = setTimeout(() => {
        performSearch(query);
    }, 300);
});

btnSearchClear.addEventListener('click', () => {
    searchInput.value = '';
    btnSearchClear.classList.add('hidden');
    searchResultsContainer.classList.add('hidden');
    treeContainer.classList.remove('hidden');
    searchInput.focus();
});

async function performSearch(query) {
    try {
        const response = await fetch(`/api/search?q=${encodeURIComponent(query)}`);
        if (!response.ok) {
            throw new Error(await response.text());
        }
        const results = await response.json();

        renderSearchResults(results, query);
    } catch (err) {
        console.error('Search failed:', err);
        searchResultsContainer.innerHTML = `
            <div style="padding: 6px 12px; font-size: 0.85rem; color: #cf222e; font-style: italic;">
                Search error: ${escapeHtml(err.message)}
            </div>
        `;
        searchResultsContainer.classList.remove('hidden');
        treeContainer.classList.add('hidden');
    }
}

function renderSearchResults(results, query) {
    searchResultsContainer.innerHTML = '';
    searchResultsContainer.classList.remove('hidden');
    treeContainer.classList.add('hidden');

    if (results.length === 0) {
        const noResultsDiv = document.createElement('div');
        noResultsDiv.style.padding = '12px';
        noResultsDiv.style.fontSize = '0.85rem';
        noResultsDiv.style.color = 'var(--text-muted)';
        noResultsDiv.style.fontStyle = 'italic';
        noResultsDiv.innerText = 'No matching markdown files found.';
        searchResultsContainer.appendChild(noResultsDiv);
        return;
    }

    results.forEach(item => {
        const node = document.createElement('div');
        node.className = 'tree-node';
        node.style.marginBottom = '12px';

        const row = document.createElement('a');
        row.className = 'tree-row';
        row.href = '#';
        row.setAttribute('data-path', item.path);

        // MD File icon
        const iconSpan = document.createElement('span');
        iconSpan.className = 'icon';
        iconSpan.innerHTML = mdIconSvg;
        iconSpan.style.color = 'var(--primary-color)';
        row.appendChild(iconSpan);

        // File name & relative path
        const textContainer = document.createElement('div');
        textContainer.style.display = 'flex';
        textContainer.style.flexDirection = 'column';
        textContainer.style.overflow = 'hidden';

        const nameLabel = document.createElement('span');
        nameLabel.style.fontWeight = '600';
        nameLabel.innerText = item.name;
        textContainer.appendChild(nameLabel);

        const pathLabel = document.createElement('span');
        pathLabel.style.fontSize = '0.75rem';
        pathLabel.style.color = 'var(--text-muted)';
        pathLabel.style.textOverflow = 'ellipsis';
        pathLabel.style.whiteSpace = 'nowrap';
        pathLabel.style.overflow = 'hidden';
        pathLabel.innerText = item.path;
        textContainer.appendChild(pathLabel);

        row.appendChild(textContainer);
        node.appendChild(row);

        // Render Snippets
        if (item.snippets && item.snippets.length > 0) {
            const snippetsContainer = document.createElement('div');
            snippetsContainer.style.marginLeft = '22px';
            snippetsContainer.style.marginTop = '4px';
            snippetsContainer.style.borderLeft = '1px solid var(--border-color)';
            snippetsContainer.style.paddingLeft = '8px';
            snippetsContainer.style.display = 'flex';
            snippetsContainer.style.flexDirection = 'column';
            snippetsContainer.style.gap = '4px';

            item.snippets.forEach((snippet, snippetIdx) => {
                const snippetDiv = document.createElement('div');
                snippetDiv.className = 'search-snippet-row';
                snippetDiv.style.fontSize = '0.75rem';
                snippetDiv.style.color = 'var(--text-muted)';
                snippetDiv.style.fontFamily = 'monospace';
                snippetDiv.style.whiteSpace = 'nowrap';
                snippetDiv.style.overflow = 'hidden';
                snippetDiv.style.textOverflow = 'ellipsis';
                snippetDiv.style.cursor = 'pointer';
                snippetDiv.style.padding = '2px 4px';
                snippetDiv.style.borderRadius = '4px';
                snippetDiv.style.transition = 'background-color 0.1s';

                // Hover effect
                snippetDiv.addEventListener('mouseenter', () => {
                    snippetDiv.style.backgroundColor = 'var(--hover-bg)';
                    snippetDiv.style.color = 'var(--text-color)';
                });
                snippetDiv.addEventListener('mouseleave', () => {
                    snippetDiv.style.backgroundColor = 'transparent';
                    snippetDiv.style.color = 'var(--text-muted)';
                });

                // Case-insensitive highlighted text
                const escapedSnippet = escapeHtml(snippet.text);
                const escapedQuery = escapeHtml(query);
                const regex = new RegExp(`(${escapedQuery.replace(/[-\/\\^$*+?.()|[\]{}]/g, '\\$&')})`, 'gi');
                const highlighted = escapedSnippet.replace(regex, '<mark style="background-color: #f8e3a1; color: #111; border-radius: 2px; padding: 0 2px;">$1</mark>');

                snippetDiv.innerHTML = `<span style="color: var(--primary-color); margin-right: 6px; font-weight: 600;">${snippet.line}:</span>${highlighted}`;

                // Clicking a specific snippet opens the file and scrolls to that snippet's match index!
                snippetDiv.addEventListener('click', (e) => {
                    e.preventDefault();
                    e.stopPropagation(); // prevent triggering parent row click

                    // Highlight the corresponding file row
                    document.querySelectorAll('.tree-row').forEach(r => r.classList.remove('active'));
                    row.classList.add('active');

                    viewFile(item.path, query, snippetIdx);
                });

                snippetsContainer.appendChild(snippetDiv);
            });

            node.appendChild(snippetsContainer);
        }

        // File click handler
        row.addEventListener('click', (e) => {
            e.preventDefault();
            // Highlight selected row in both trees
            document.querySelectorAll('.tree-row').forEach(r => r.classList.remove('active'));
            row.classList.add('active');

            viewFile(item.path, query);
        });

        searchResultsContainer.appendChild(node);
    });
}

function updateMarkdownDisplay() {
    if (!currentFileData) return;

    const btnText = btnToggleSource.querySelector('span');

    if (showingSource) {
        renderSourceCode(currentFileData.content, currentFileData.ext);
        btnText.innerText = 'View Rendered';
    } else {
        renderParsedMarkdown(currentFileData.html, currentFileData.path);
        btnText.innerText = 'View Source';
    }

    // Keep matches highlighted when toggling modes
    const query = searchInput.value.trim();
    if (query !== '' && !searchResultsContainer.classList.contains('hidden')) {
        highlightAndScrollToQuery(query);
    }
}

// Set up the auto-refresh interval based on the checkbox state
function setupAutofresh() {
    if (autofreshIntervalId) {
        clearInterval(autofreshIntervalId);
        autofreshIntervalId = null;
    }

    if (toggleAutofreshCheckbox.checked) {
        autofreshIntervalId = setInterval(() => {
            // Only refresh if the browser tab is currently active to save server resource
            if (!document.hidden) {
                refreshTreeSilent();
            }
        }, refreshIntervalSec * 1000);
    }
}

// SVG Icons helper
const folderIconSvg = `
    <svg viewBox="0 0 16 16" width="16" height="16" fill="currentColor">
        <path d="M1.75 1A1.75 1.75 0 0 0 0 2.75v10.5C0 14.216.784 15 1.75 15h12.5A1.75 1.75 0 0 0 16 13.25v-8.5A1.75 1.75 0 0 0 14.25 3H7.5a.25.25 0 0 1-.2-.1l-.9-1.2C6.07 1.28 5.55 1 5 1H1.75z"></path>
    </svg>
`;

const fileIconSvg = `
    <svg viewBox="0 0 16 16" width="16" height="16" fill="currentColor">
        <path d="M2 1.75C2 .784 2.784 0 3.75 0h6.586c.464 0 .909.184 1.237.513l3.25 3.25c.329.329.513.773.513 1.237v9.25A1.75 1.75 0 0 1 13.5 16h-9.75A1.75 1.75 0 0 1 2 14.25V1.75zM3.75 1.5a.25.25 0 0 0-.25.25v12.5c0 .138.112.25.25.25h9.75a.25.25 0 0 0 .25-.25V6h-3.25A1.75 1.75 0 0 1 9 4.25V1.5H3.75zM10.5 1.811V4.25c0 .138.112.25.25.25h2.439L10.5 1.811z"></path>
    </svg>
`;

const mdIconSvg = `
    <svg viewBox="0 0 16 16" width="16" height="16" fill="currentColor">
        <path d="M14.854 4.854a.5.5 0 0 0 0-.708l-4-4a.5.5 0 0 0-.708 0L6.293 4H1.5A1.5 1.5 0 0 0 0 5.5v9A1.5 1.5 0 0 0 1.5 16h13a1.5 1.5 0 0 0 1.5-1.5V5a.5.5 0 0 0-.146-.346zM10.5 1.707L13.793 5H10.5V1.707zM1.5 5h8v1a.5.5 0 0 0 .5.5h3.5v8a.5.5 0 0 1-.5.5h-11a.5.5 0 0 1-.5-.5v-9A.5.5 0 0 1 1.5 5zm2.5 3h1v4h-1V8zm2.5 0H8v1.5l1-1.5h1.5L9 9.5 10.5 12H9l-1-1.5V12H6.5V8z"/>
    </svg>
`;

const arrowIconSvg = `
    <svg viewBox="0 0 16 16" width="12" height="12" fill="currentColor" class="icon-arrow">
        <path d="M6.22 3.22a.75.75 0 0 1 1.06 0l4.25 4.25a.75.75 0 0 1 0 1.06l-4.25 4.25a.75.75 0 0 1-1.06-1.06L9.94 8 6.22 4.28a.75.75 0 0 1 0-1.06z"></path>
    </svg>
`;

// Initialize and load the root tree
async function init() {
    try {
        const res = await fetch('/api/config');
        if (res.ok) {
            const config = await res.json();
            if (config.refresh_interval) {
                refreshIntervalSec = config.refresh_interval;
            }
        }
    } catch (e) {
        console.error("Failed to load config:", e);
    }
    await loadDirectory('', treeContainer, true);
    setupAutofresh();
}

async function refreshTree() {
    treeContainer.innerHTML = '';
    await loadDirectory('', treeContainer, true);
}

// butter-smooth, completely flicker-free silent reload of directory tree structure
async function refreshTreeSilent() {
    await reconcileDirectory('', treeContainer);
}

// Factory function to create a new folder or file tree node
function createTreeNode(item) {
    const node = document.createElement('div');
    node.className = 'tree-node';

    const row = document.createElement('a');
    row.className = 'tree-row';
    row.href = '#';
    row.setAttribute('data-path', item.path);
    row.setAttribute('data-isdir', item.is_dir);

    // Render Arrow Icon if directory
    const arrowSpan = document.createElement('span');
    arrowSpan.className = 'icon';
    if (item.is_dir) {
        arrowSpan.innerHTML = arrowIconSvg;
    }
    row.appendChild(arrowSpan);

    // Render Type Icon (Folder, MD File, or normal File)
    const typeSpan = document.createElement('span');
    typeSpan.className = 'icon';
    if (item.is_dir) {
        typeSpan.innerHTML = folderIconSvg;
    } else if (item.name.toLowerCase().endsWith('.md') || item.name.toLowerCase().endsWith('.markdown')) {
        typeSpan.innerHTML = mdIconSvg;
        typeSpan.style.color = 'var(--primary-color)';
    } else {
        typeSpan.innerHTML = fileIconSvg;
    }
    row.appendChild(typeSpan);

    // Name label
    const nameLabel = document.createElement('span');
    nameLabel.innerText = item.name;
    row.appendChild(nameLabel);

    if (item.is_dir && !toggleAllCheckbox.checked && item.has_no_md) {
        const emptyLabel = document.createElement('span');
        emptyLabel.className = 'no-md-label';
        emptyLabel.style.color = 'var(--text-muted)';
        emptyLabel.style.fontSize = '0.8rem';
        emptyLabel.style.fontStyle = 'italic';
        emptyLabel.style.marginLeft = '6px';
        emptyLabel.innerText = '(no md)';
        row.appendChild(emptyLabel);
    }

    node.appendChild(row);

    // If folder, add children container
    if (item.is_dir) {
        const childrenContainer = document.createElement('div');
        childrenContainer.className = 'tree-children';
        node.appendChild(childrenContainer);

        row.addEventListener('click', (e) => {
            e.preventDefault();
            const arrowSvg = arrowSpan.querySelector('.icon-arrow');
            
            if (childrenContainer.classList.contains('expanded')) {
                childrenContainer.classList.remove('expanded');
                if (arrowSvg) arrowSvg.classList.remove('open');
                expandedPaths.delete(item.path);
            } else {
                childrenContainer.classList.add('expanded');
                if (arrowSvg) arrowSvg.classList.add('open');
                expandedPaths.add(item.path);
                
                // Lazy load contents if not already loaded
                if (childrenContainer.children.length === 0) {
                    loadDirectory(item.path, childrenContainer);
                }
            }
        });
    } else {
        // File click handler
        row.addEventListener('click', (e) => {
            e.preventDefault();
            
            // Highlight clicked row
            document.querySelectorAll('.tree-row').forEach(r => r.classList.remove('active'));
            row.classList.add('active');

            viewFile(item.path);
        });
    }

    return node;
}

// In-place DOM reconciliation (DOM patching) of the directory elements to eliminate flickering
async function reconcileDirectory(path, container) {
    try {
        const showAll = toggleAllCheckbox.checked;
        const response = await fetch(`/api/list?path=${encodeURIComponent(path)}&all=${showAll}`);
        if (!response.ok) {
            throw new Error(await response.text());
        }
        const newItems = await response.json();

        // Sort items: directories first, then alphabetically
        newItems.sort((a, b) => {
            if (a.is_dir !== b.is_dir) {
                return a.is_dir ? -1 : 1;
            }
            return a.name.localeCompare(b.name);
        });

        // Get existing DOM nodes inside this container
        const existingNodes = Array.from(container.children).filter(el => el.classList.contains('tree-node'));
        const existingNodesMap = new Map();
        existingNodes.forEach(node => {
            const row = node.querySelector('.tree-row');
            if (row) {
                const itemPath = row.getAttribute('data-path');
                existingNodesMap.set(itemPath, node);
            }
        });

        // Clear any (empty) or (error) helper divs in the container
        Array.from(container.children).forEach(el => {
            if (!el.classList.contains('tree-node')) {
                container.removeChild(el);
            }
        });

        // If no new items, display (empty) and remove all existing nodes
        if (newItems.length === 0) {
            container.innerHTML = '';
            if (showAll || container === treeContainer) {
                const emptyDiv = document.createElement('div');
                emptyDiv.style.padding = '6px 12px';
                emptyDiv.style.fontSize = '0.85rem';
                emptyDiv.style.color = 'var(--text-muted)';
                emptyDiv.style.fontStyle = 'italic';
                emptyDiv.innerText = '(empty)';
                container.appendChild(emptyDiv);
            }
            return;
        }

        const activeRow = document.querySelector('.tree-row.active');
        const activePath = activeRow ? activeRow.getAttribute('data-path') : '';

        newItems.forEach((item, index) => {
            let node = existingNodesMap.get(item.path);

            if (node) {
                // Node already exists! Keep it and remove from map so we know it's still alive
                existingNodesMap.delete(item.path);

                // If highlighted, restore highlight
                if (activePath && item.path === activePath) {
                    const row = node.querySelector('.tree-row');
                    if (row) row.classList.add('active');
                }

                // Update empty-label dynamically
                if (item.is_dir) {
                    const row = node.querySelector('.tree-row');
                    if (row) {
                        let emptyLabel = row.querySelector('.no-md-label');
                        const showEmpty = !toggleAllCheckbox.checked && item.has_no_md;
                        if (showEmpty) {
                            if (!emptyLabel) {
                                emptyLabel = document.createElement('span');
                                emptyLabel.className = 'no-md-label';
                                emptyLabel.style.color = 'var(--text-muted)';
                                emptyLabel.style.fontSize = '0.8rem';
                                emptyLabel.style.fontStyle = 'italic';
                                emptyLabel.style.marginLeft = '6px';
                                emptyLabel.innerText = '(no md)';
                                row.appendChild(emptyLabel);
                            }
                        } else {
                            if (emptyLabel) {
                                row.removeChild(emptyLabel);
                            }
                        }
                    }
                }

                // If it's a directory and it is expanded, recursively reconcile its children
                if (item.is_dir && expandedPaths.has(item.path)) {
                    const childrenContainer = node.querySelector('.tree-children');
                    if (childrenContainer) {
                        reconcileDirectory(item.path, childrenContainer);
                    }
                }
            } else {
                // Node does not exist! Create a new tree node
                node = createTreeNode(item);
            }

            // Insert or re-order node in container at index position
            const currentChildAtIndex = container.children[index];
            if (currentChildAtIndex !== node) {
                container.insertBefore(node, currentChildAtIndex || null);
            }
        });

        // Any nodes left in existingNodesMap are stale (deleted on disk) - remove them!
        existingNodesMap.forEach(node => {
            container.removeChild(node);
        });

        // If container became empty, add (empty) label
        if (container.children.length === 0) {
            if (showAll || container === treeContainer) {
                const emptyDiv = document.createElement('div');
                emptyDiv.style.padding = '6px 12px';
                emptyDiv.style.fontSize = '0.85rem';
                emptyDiv.style.color = 'var(--text-muted)';
                emptyDiv.style.fontStyle = 'italic';
                emptyDiv.innerText = '(empty)';
                container.appendChild(emptyDiv);
            }
        }
    } catch (err) {
        console.error('Error reconciling directory:', err);
    }
}

// Load a directory path from the server and append list items to container
async function loadDirectory(path, container, autoExpand = false) {
    try {
        const showAll = toggleAllCheckbox.checked;
        const response = await fetch(`/api/list?path=${encodeURIComponent(path)}&all=${showAll}`);
        if (!response.ok) {
            throw new Error(await response.text());
        }
        const items = await response.json();
        
        // Sort items: directories first, then alphabetically
        items.sort((a, b) => {
            if (a.is_dir !== b.is_dir) {
                return a.is_dir ? -1 : 1;
            }
            return a.name.localeCompare(b.name);
        });

        if (items.length === 0) {
            if (showAll || container === treeContainer) {
                const emptyDiv = document.createElement('div');
                emptyDiv.style.padding = '6px 12px';
                emptyDiv.style.fontSize = '0.85rem';
                emptyDiv.style.color = 'var(--text-muted)';
                emptyDiv.style.fontStyle = 'italic';
                emptyDiv.innerText = '(empty)';
                container.appendChild(emptyDiv);
            }
            return;
        }

        const foldersToExpand = [];

        items.forEach(item => {
            const node = createTreeNode(item);
            container.appendChild(node);

            if (item.is_dir) {
                const childrenContainer = node.querySelector('.tree-children');
                const arrowSpan = node.querySelector('.icon'); // first icon is arrow span

                // Restore previous expanded state during refresh
                if (expandedPaths.has(item.path)) {
                    childrenContainer.classList.add('expanded');
                    const arrowSvg = arrowSpan.querySelector('.icon-arrow');
                    if (arrowSvg) arrowSvg.classList.add('open');
                    loadDirectory(item.path, childrenContainer);
                } else if (autoExpand) {
                    foldersToExpand.push({ childrenContainer, arrowSpan, path: item.path });
                }
            }
        });

        // Auto-expand if requested and fits the screen
        if (autoExpand && foldersToExpand.length > 0) {
            for (const folder of foldersToExpand) {
                // Check if current elements already cause scrolling or exceed threshold
                const doesFit = treeContainer.scrollHeight <= treeContainer.clientHeight;
                const visibleRows = treeContainer.querySelectorAll('.tree-row').length;

                if (doesFit && visibleRows < 40) {
                    folder.childrenContainer.classList.add('expanded');
                    const arrowSvg = folder.arrowSpan.querySelector('.icon-arrow');
                    if (arrowSvg) arrowSvg.classList.add('open');
                    expandedPaths.add(folder.path); // Track in expandedPaths
                    
                    await loadDirectory(folder.path, folder.childrenContainer, true);
                } else {
                    break;
                }
            }
        }
    } catch (err) {
        console.error('Error loading directory:', err);
        const errDiv = document.createElement('div');
        errDiv.style.padding = '6px 12px';
        errDiv.style.fontSize = '0.85rem';
        errDiv.style.color = '#cf222e'; // error color
        errDiv.innerText = `Error: ${err.message}`;
        container.appendChild(errDiv);
    }
}

async function expandAllNode(container) {
    const rows = container.querySelectorAll('.tree-row[data-isdir="true"]');
    for (const row of rows) {
        const childrenContainer = row.nextElementSibling;
        const arrowSpan = row.querySelector('.icon');
        const arrowSvg = arrowSpan ? arrowSpan.querySelector('.icon-arrow') : null;
        if (childrenContainer && !childrenContainer.classList.contains('expanded')) {
            childrenContainer.classList.add('expanded');
            if (arrowSvg) arrowSvg.classList.add('open');
            
            if (childrenContainer.children.length === 0) {
                const path = row.getAttribute('data-path');
                await loadDirectory(path, childrenContainer);
            }
            // Recursively expand newly loaded children
            await expandAllNode(childrenContainer);
        }
    }
}

// Helper to format file sizes
function formatBytes(bytes) {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// Helper to escape HTML to prevent XSS
function escapeHtml(str) {
    return str
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#039;');
}

// Helper to get parent directory of a path
function getParentDir(path) {
    if (!path) return '';
    const lastSlash = path.lastIndexOf('/');
    if (lastSlash === -1) return '';
    return path.substring(0, lastSlash);
}

// Helper to resolve relative path from baseDir
function resolveRelativePath(baseDir, relPath) {
    if (relPath.startsWith('/')) {
        relPath = relPath.substring(1);
        baseDir = '';
    }
    const parts = (baseDir ? baseDir.split('/') : []).concat(relPath.split('/'));
    const resolved = [];
    for (const part of parts) {
        if (part === '' || part === '.') {
            continue;
        }
        if (part === '..') {
            if (resolved.length > 0) {
                resolved.pop();
            }
        } else {
            resolved.push(part);
        }
    }
    return resolved.join('/');
}

// Helper to identify if URL is external
function isExternalUrl(url) {
    return /^(https?:|data:|mailto:|tel:)/i.test(url);
}

function highlightAndScrollToQuery(query, matchIndex = 0) {
    if (!query) return;
    const queryLower = query.toLowerCase();

    // Clean up any stale highlights from previously viewed files
    removeHighlights(markdownContent);

    const walker = document.createTreeWalker(
        markdownContent,
        NodeFilter.SHOW_TEXT,
        null,
        false
    );

    const nodesToReplace = [];
    let currentNode = walker.nextNode();
    while (currentNode) {
        const parent = currentNode.parentNode;
        const parentTag = parent.tagName.toUpperCase();
        if (parentTag !== 'SCRIPT' && parentTag !== 'STYLE') {
            const text = currentNode.nodeValue;
            if (text.toLowerCase().includes(queryLower)) {
                nodesToReplace.push(currentNode);
            }
        }
        currentNode = walker.nextNode();
    }

    nodesToReplace.forEach(node => {
        const parent = node.parentNode;
        const text = node.nodeValue;
        const regex = new RegExp(`(${query.replace(/[-\/\\^$*+?.()|[\]{}]/g, '\\$&')})`, 'gi');
        const parts = text.split(regex);

        const fragment = document.createDocumentFragment();
        parts.forEach(part => {
            if (part.toLowerCase() === queryLower) {
                const mark = document.createElement('mark');
                mark.className = 'search-highlight';
                mark.style.backgroundColor = '#f8e3a1';
                mark.style.color = '#111';
                mark.style.borderRadius = '2px';
                mark.style.padding = '0 2px';
                mark.innerText = part;
                fragment.appendChild(mark);
            } else {
                fragment.appendChild(document.createTextNode(part));
            }
        });

        parent.replaceChild(fragment, node);
    });

    // Smoothly scroll and center the specified match index in the viewport
    const allMarks = markdownContent.querySelectorAll('mark.search-highlight');
    const targetMark = allMarks[matchIndex] || allMarks[0];
    if (targetMark) {
        setTimeout(() => {
            targetMark.scrollIntoView({ behavior: 'smooth', block: 'center' });
        }, 100);
    }
}

function removeHighlights(container) {
    const marks = container.querySelectorAll('mark.search-highlight');
    marks.forEach(mark => {
        const parent = mark.parentNode;
        parent.replaceChild(document.createTextNode(mark.innerText), mark);
        parent.normalize(); // Cleanly merges adjacent text nodes
    });
}

function collapseAll() {
    const children = treeContainer.querySelectorAll('.tree-children.expanded');
    children.forEach(el => {
        el.classList.remove('expanded');
    });
    const arrows = treeContainer.querySelectorAll('.icon-arrow.open');
    arrows.forEach(el => {
        el.classList.remove('open');
    });
}

function renderParsedMarkdown(html, path) {
    markdownContent.innerHTML = `<div class="markdown-body">${html}</div>`;

    // Process images inside markdown to serve local assets securely
    const docDir = getParentDir(path);
    markdownContent.querySelectorAll('.markdown-body img').forEach((img) => {
        const src = img.getAttribute('src');
        if (src && !isExternalUrl(src)) {
            const resolved = resolveRelativePath(docDir, src);
            img.src = `/api/raw?path=${encodeURIComponent(resolved)}`;
        }
    });

    // Run syntax highlighting on code blocks inside markdown
    if (window.hljs) {
        document.querySelectorAll('#markdown-content .markdown-body pre code').forEach((el) => {
            hljs.highlightElement(el);
        });
    }
}

function renderSourceCode(content, ext) {
    const lines = content.split(/\r?\n/);
    // Ensure we render at least 1 line
    const lineCount = Math.max(lines.length, 1);
    let gutterHtml = '';
    for (let i = 1; i <= lineCount; i++) {
        gutterHtml += `<div style="height: 20px; line-height: 20px;">${i}</div>`;
    }

    markdownContent.innerHTML = `
        <div style="display: flex; background: #ffffff; border-radius: 8px; border: 1px solid #d0d7de; font-family: ui-monospace, SFMono-Regular, SF Mono, Menlo, Consolas, Liberation Mono, monospace; font-size: 13px; overflow: hidden; box-shadow: 0 4px 24px rgba(0, 0, 0, 0.45);">
            <div class="line-gutter" style="padding: 16px 12px; text-align: right; color: #57606a; background-color: #f6f8fa; border-right: 1px solid #d0d7de; user-select: none; min-width: 48px; flex-shrink: 0; font-size: 13px; font-family: inherit;">
                ${gutterHtml}
            </div>
            <pre style="margin: 0; padding: 16px; overflow-x: auto; flex: 1; background-color: #ffffff; color: #1f2328; font-family: inherit;"><code class="language-${ext}" style="background-color: #ffffff; color: #1f2328; padding: 0; display: block; white-space: pre; line-height: 20px; font-family: inherit; font-size: 13px;">${escapeHtml(content)}</code></pre>
        </div>
    `;

    // Run syntax highlighting on the new block
    if (window.hljs) {
        const el = markdownContent.querySelector('pre code');
        if (el) hljs.highlightElement(el);
    }
}

// View rendering for files (text/markdown or binary files)
async function viewFile(path, searchQuery = '', matchIndex = 0) {
    try {
        const response = await fetch(`/api/view?path=${encodeURIComponent(path)}`);
        if (!response.ok) {
            throw new Error(await response.text());
        }
        const data = await response.json();

        currentActivePath = data.path;
        currentFileData = data;
        showingSource = false; // Always default to rendered view when opening a new file
        filePathDisplay.innerText = data.path;

        if (data.is_binary) {
            btnToggleSource.classList.add('hidden');
            // Display details of the binary file without parsing it
            markdownContent.innerHTML = `
                <div style="display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 48px 24px; text-align: center;">
                    <svg viewBox="0 0 24 24" width="72" height="72" style="color: #afb8c1; margin-bottom: 20px;" fill="currentColor">
                        <path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-1 16H6c-.55 0-1-.45-1-1V6c0-.55.45-1 1-1h12c.55 0 1 .45 1 1v12c0 .55-.45 1-1 1zm-4-4h-4v-2h4v2zm2-4H8V9h8v2z"/>
                    </svg>
                    <h3 style="margin-bottom: 8px; font-weight: 600; color: #24292f; font-size: 1.4rem;">${escapeHtml(data.title)}</h3>
                    <p style="font-size: 0.95rem; margin-bottom: 24px; color: #57606a;">Size: ${formatBytes(data.size)}</p>
                    <div style="display: flex; flex-direction: column; align-items: center; gap: 8px;">
                        <span style="font-size: 0.85rem; padding: 6px 16px; background-color: #f6f8fa; border: 1px solid #d0d7de; border-radius: 20px; color: #24292f; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em;">Binary File</span>
                        <p style="font-size: 0.85rem; color: #57606a; max-width: 320px; margin-top: 12px; line-height: 1.4;">This file contains binary content and cannot be viewed as a parsed document.</p>
                    </div>
                </div>
            `;
        } else if (data.is_markdown) {
            btnToggleSource.classList.remove('hidden');
            updateMarkdownDisplay();
            if (searchQuery) {
                highlightAndScrollToQuery(searchQuery, matchIndex);
            }
        } else {
            btnToggleSource.classList.add('hidden');
            renderSourceCode(data.content, data.ext);
            if (searchQuery) {
                highlightAndScrollToQuery(searchQuery, matchIndex);
            }
        }

        // Toggle visibility
        placeholderView.classList.add('hidden');
        markdownView.classList.remove('hidden');

        // Scroll view-header back to top only if not focusing on a query match
        if (!searchQuery) {
            document.querySelector('.content-scroller').scrollTop = 0;
        }
    } catch (err) {
        alert(`Failed to load file: ${err.message}`);
    }
}

init();