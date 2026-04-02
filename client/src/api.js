// client/src/api.js
const API_BASE_URL = 'http://localhost:8080/v1';

// 1. DELETE the hardcoded API_KEY variable completely!

// 2. Change 'headers' into a FUNCTION that grabs the key right when it's needed
const getHeaders = () => {
  const apiKey = localStorage.getItem('rss_api_key');
  
  return {
    'Content-Type': 'application/json',
    // If there is no key (like before logging in), it just passes "ApiKey null" 
    // which your Go backend will correctly reject until they log in!
    'Authorization': `ApiKey ${apiKey}`
  };
};

export const fetchPosts = async (page = 1) => {
  // Now it attaches ?page=X to the URL!
  const response = await fetch(`${API_BASE_URL}/posts?page=${page}`, { headers: getHeaders() });
  if (!response.ok) throw new Error('Failed to fetch posts');
  return response.json();
};

export const searchPosts = async (searchTerm) => {
  if (!searchTerm.trim()) {
    return fetchPosts(); 
  }
  
  // Update here too
  const response = await fetch(`${API_BASE_URL}/posts/search?q=${encodeURIComponent(searchTerm)}`, { headers: getHeaders() });
  if (!response.ok) throw new Error('Search failed');
  return response.json();
};

export const addFeed = async (name, url) => {
  const response = await fetch(`${API_BASE_URL}/feeds`, {
    method: 'POST',
    headers: getHeaders(), // Update here too
    body: JSON.stringify({ name, url }),
  });
  
  if (!response.ok) {
    throw new Error('Failed to add feed. It might already exist or the URL is invalid.');
  }
  
  return response.json();
};

export const fetchFeeds = async () => {
  // Update here too
  const response = await fetch(`${API_BASE_URL}/feeds`, { headers: getHeaders() });
  if (!response.ok) throw new Error('Failed to fetch feeds');
  return response.json();
};

export const fetchFeedFollows = async () => {
  const response = await fetch(`${API_BASE_URL}/feed_follows`, { headers: getHeaders() });
  // If the endpoint doesn't exist or errors out, we just return an empty array safely
  if (!response.ok) return []; 
  return response.json();
};

export const unfollowFeed = async (feedId) => {
  const response = await fetch(`${API_BASE_URL}/feed_follows/${feedId}`, {
    method: 'DELETE',
    headers: getHeaders(), // Grabs your auth key automatically!
  });
  if (!response.ok) throw new Error('Failed to unfollow feed');
  return response.json(); 
};