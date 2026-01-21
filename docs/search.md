---
layout: default
title: Search
description: Search the Yougopher documentation
---

<div class="search-container">
  <input type="text" id="search-input" class="search-input" placeholder="Search documentation..." autofocus>
  <div id="search-results" class="search-results"></div>
</div>

<script src="https://unpkg.com/lunr/lunr.js"></script>
<script>
(function() {
  let searchIndex = null;
  let searchData = null;

  // Load the search index
  fetch('{{ "/search.json" | relative_url }}')
    .then(response => response.json())
    .then(data => {
      searchData = data;
      searchIndex = lunr(function() {
        this.ref('url');
        this.field('title', { boost: 10 });
        this.field('description', { boost: 5 });
        this.field('content');

        data.forEach(doc => {
          this.add(doc);
        });
      });
    });

  const searchInput = document.getElementById('search-input');
  const searchResults = document.getElementById('search-results');

  // Get query from URL if present
  const urlParams = new URLSearchParams(window.location.search);
  const query = urlParams.get('q');
  if (query) {
    searchInput.value = query;
    setTimeout(() => performSearch(query), 100);
  }

  searchInput.addEventListener('input', function() {
    const query = this.value.trim();
    if (query.length < 2) {
      searchResults.innerHTML = '<p class="search-hint">Type at least 2 characters to search...</p>';
      return;
    }
    performSearch(query);
  });

  function performSearch(query) {
    if (!searchIndex) {
      searchResults.innerHTML = '<p>Loading search index...</p>';
      return;
    }

    try {
      // Add wildcards for partial matching
      const searchTerms = query.split(' ').map(term => term + '*').join(' ');
      const results = searchIndex.search(searchTerms);

      if (results.length === 0) {
        searchResults.innerHTML = '<p class="no-results">No results found for "' + escapeHtml(query) + '"</p>';
        return;
      }

      let html = '<ul class="results-list">';
      results.slice(0, 20).forEach(result => {
        const doc = searchData.find(d => d.url === result.ref);
        if (doc) {
          const snippet = getSnippet(doc.content, query);
          html += `
            <li class="result-item">
              <a href="${doc.url}" class="result-title">${escapeHtml(doc.title)}</a>
              ${doc.description ? `<p class="result-description">${escapeHtml(doc.description)}</p>` : ''}
              <p class="result-snippet">${snippet}</p>
            </li>
          `;
        }
      });
      html += '</ul>';
      searchResults.innerHTML = html;
    } catch (e) {
      // Fallback to simple search if lunr query fails
      const results = searchData.filter(doc =>
        doc.title.toLowerCase().includes(query.toLowerCase()) ||
        doc.content.toLowerCase().includes(query.toLowerCase())
      );

      if (results.length === 0) {
        searchResults.innerHTML = '<p class="no-results">No results found for "' + escapeHtml(query) + '"</p>';
        return;
      }

      let html = '<ul class="results-list">';
      results.slice(0, 20).forEach(doc => {
        const snippet = getSnippet(doc.content, query);
        html += `
          <li class="result-item">
            <a href="${doc.url}" class="result-title">${escapeHtml(doc.title)}</a>
            ${doc.description ? `<p class="result-description">${escapeHtml(doc.description)}</p>` : ''}
            <p class="result-snippet">${snippet}</p>
          </li>
        `;
      });
      html += '</ul>';
      searchResults.innerHTML = html;
    }
  }

  function getSnippet(content, query) {
    const lowerContent = content.toLowerCase();
    const lowerQuery = query.toLowerCase().split(' ')[0];
    const index = lowerContent.indexOf(lowerQuery);

    if (index === -1) {
      return escapeHtml(content.substring(0, 150)) + '...';
    }

    const start = Math.max(0, index - 50);
    const end = Math.min(content.length, index + 100);
    let snippet = content.substring(start, end);

    if (start > 0) snippet = '...' + snippet;
    if (end < content.length) snippet = snippet + '...';

    return escapeHtml(snippet);
  }

  function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }
})();
</script>
