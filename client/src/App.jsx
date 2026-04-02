import { fetchPosts, searchPosts, fetchFeeds, fetchFeedFollows, unfollowFeed } from './api';
import React, { useState, useEffect } from 'react';
import AddFeedModal from './AddFeedModal';

// --- NEW COMPONENT: The Login / Sign Up Screen ---
const LoginScreen = ({ onLogin }) => {
  const [isLoginMode, setIsLoginMode] = useState(true);
  const [inputValue, setInputValue] = useState('');
  const [error, setError] = useState('');
  
  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');

    if (isLoginMode) {
      // LOG IN: Just save the pasted API key
      if (inputValue.trim().length < 10) {
        setError('Please enter a valid API key.');
        return;
      }
      onLogin(inputValue.trim());
    } else {
      // SIGN UP: Hit the Go backend to create a new user
      if (inputValue.trim().length < 2) {
        setError('Please enter a valid name.');
        return;
      }
      try {
        const response = await fetch('http://localhost:8080/v1/users', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ name: inputValue.trim() })
        });
        
        if (!response.ok) throw new Error('Failed to create account');
        
        const data = await response.json();
        // The backend returns the new user object, which includes their new API key!
        onLogin(data.api_key); 
      } catch (err) {
        setError('Could not create account. Is the server running?');
      }
    }
  };

  return (
    <div className="min-h-screen bg-gray-950 flex items-center justify-center p-4 text-gray-100 font-sans">
      <div className="bg-gray-900 border border-gray-800 p-8 rounded-xl shadow-2xl w-full max-w-md">
        <h1 className="text-3xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-emerald-400 to-cyan-500 mb-6 text-center">
          Welcome to CoreDigest
        </h1>
        
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-400 mb-1">
              {isLoginMode ? 'Enter your API Key' : 'Enter your Name'}
            </label>
            <input 
              type="text" 
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              placeholder={isLoginMode ? "Paste your 64-character key..." : "e.g. John Doe"}
              className="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 focus:outline-none focus:ring-2 focus:ring-emerald-500 text-gray-100"
            />
          </div>
          
          {error && <p className="text-red-400 text-sm text-center">{error}</p>}
          
          <button type="submit" className="w-full bg-emerald-600 hover:bg-emerald-500 text-white font-bold py-3 rounded-lg transition-colors">
            {isLoginMode ? 'Access Dashboard' : 'Generate API Key'}
          </button>
        </form>

        <div className="mt-6 text-center text-sm text-gray-500">
          {isLoginMode ? "Don't have an account? " : "Already have a key? "}
          <button 
            onClick={() => { setIsLoginMode(!isLoginMode); setInputValue(''); setError(''); }}
            className="text-emerald-400 hover:underline focus:outline-none"
          >
            {isLoginMode ? "Create one" : "Log in"}
          </button>
        </div>
      </div>
    </div>
  );
};

// --- YOUR MAIN APP COMPONENT ---
function App() {
  // 1. ADD AUTHENTICATION STATE (Read from browser memory on load)
  const [apiKey, setApiKey] = useState(localStorage.getItem('rss_api_key') || null);

  const [posts, setPosts] = useState([]);
  const [feeds, setFeeds] = useState([]); 
  const [activeFeedId, setActiveFeedId] = useState(null); 
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [isSearching, setIsSearching] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true); 
  const [isLoadingMore, setIsLoadingMore] = useState(false);

  // ---> NEW: Tracks which AI summaries are expanded
  const [expandedSummaries, setExpandedSummaries] = useState({});

  // Helper function to flip the toggle for a specific post
  const toggleSummary = (postId) => {
    setExpandedSummaries(prev => ({
      ...prev,
      [postId]: !prev[postId]
    }));
  };

  // 2. ONLY LOAD DATA IF API KEY EXISTS
  const loadDashboardData = async () => {
    if (!apiKey) return; 

    setLoading(true);
    try {
      // 1. Fetch EVERYTHING at the same time for maximum speed
      const [postsData, globalFeedsData, followsData] = await Promise.all([
        fetchPosts(),
        fetchFeeds(),
        fetchFeedFollows()
      ]);

      // 2. Extract just the IDs of the feeds this specific user follows
      // (Safety check: Make sure followsData is an array before mapping)
      const followedFeedIds = Array.isArray(followsData) 
        ? followsData.map(follow => follow.feed_id) 
        : [];

      // 3. Filter the global feeds to ONLY include the ones the user follows!
      const myPersonalFeeds = (globalFeedsData || []).filter(feed => 
        followedFeedIds.includes(feed.id)
      );

      setPosts(postsData || []);
      setFeeds(myPersonalFeeds); // The sidebar is now personalized!
      setError(null);
    } catch (err) {
      setError('Could not connect to the Go backend.');
    } finally {
      setLoading(false);
    }
  };

  const loadMorePosts = async () => {
    if (isLoadingMore || !hasMore) return;
    
    setIsLoadingMore(true);
    try {
      const nextPage = page + 1;
      const newPosts = await fetchPosts(nextPage);
      
      if (newPosts && newPosts.length > 0) {
        // Attach the new posts to the bottom of the existing posts array!
        setPosts(prevPosts => [...prevPosts, ...newPosts]);
        setPage(nextPage);
      } else {
        // If the backend returns an empty array, we've reached the end!
        setHasMore(false);
      }
    } catch (err) {
      console.error("Failed to load more posts", err);
    } finally {
      setIsLoadingMore(false);
    }
  };

  useEffect(() => {
    loadDashboardData();
  }, [apiKey]); // Re-run this effect whenever the apiKey changes (like when logging in)

  // 3. LOGIN & LOGOUT HANDLERS
  const handleLogin = (key) => {
    localStorage.setItem('rss_api_key', key);
    setApiKey(key);
  };

  const handleLogout = () => {
    localStorage.removeItem('rss_api_key');
    setApiKey(null);
    setPosts([]);
    setFeeds([]);
  };

  const handleUnfollow = async (feedId, e) => {
    e.stopPropagation(); // Stops the feed from turning green when you click the 'X'
    
    if (!window.confirm("Are you sure you want to unfollow this feed?")) return;

    try {
      await unfollowFeed(feedId);
      
      // If you are viewing the feed you just deleted, kick back to "All Articles"
      if (activeFeedId === feedId) {
        setActiveFeedId(null);
      }
      
      // Reset the page and reload the dashboard to clear the articles!
      setPage(1); 
      loadDashboardData();
    } catch (err) {
      alert("Failed to unfollow the feed. Please try again.");
    }
  };

  const handleCopyKey = () => {
    navigator.clipboard.writeText(apiKey);
    alert("API Key copied to clipboard! Save this somewhere safe so you can log back in later.");
  };

  const handleSearch = async (e) => {
    if (e.key === 'Enter') {
      setIsSearching(true);
      setActiveFeedId(null);
      setError(null);
      try {
        const results = await searchPosts(searchQuery);
        setPosts(results || []);
      } catch (err) {
        setError('Search failed. Please try again.');
      } finally {
        setIsSearching(false);
      }
    }
  };

  const displayedPosts = activeFeedId 
    ? posts.filter(post => post.feed_id === activeFeedId)
    : posts;

  // 4. THE GATEKEEPER: If no key, show login screen!
  if (!apiKey) {
    return <LoginScreen onLogin={handleLogin} />;
  }

  // 5. THE DASHBOARD (Your existing code, plus a Logout button)
  return (
    <div className="min-h-screen bg-gray-950 text-gray-100 font-sans relative flex flex-col">
      <nav className="sticky top-0 z-40 bg-gray-900 border-b border-gray-800 px-6 py-4 flex justify-between items-center shadow-md">
        <h1 
          className="text-2xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-emerald-400 to-cyan-500 pb-1 cursor-pointer"
          onClick={() => { setActiveFeedId(null); setSearchQuery(''); loadDashboardData(); }}
        >
          CoreDigest
        </h1>
        
        <div className="flex items-center space-x-4 w-1/2 justify-end">
          <div className="w-2/3 relative">
            <input 
              type="text" 
              placeholder="Search AI summaries... (Press Enter)" 
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyDown={handleSearch}
              className="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500 text-gray-100 placeholder-gray-500 transition-all"
            />
          </div>
          <button 
            onClick={() => setIsModalOpen(true)}
            className="bg-emerald-600 hover:bg-emerald-500 text-white px-4 py-2 rounded-lg text-sm font-semibold transition-colors whitespace-nowrap"
          >
            + Add Feed
          </button>

          {/* ---> NEW COPY KEY BUTTON <--- */}
          <button 
            onClick={handleCopyKey}
            className="text-gray-400 hover:text-emerald-400 px-2 py-2 text-sm font-medium transition-colors"
            title="Copy API Key"
          >
            Copy Key
          </button>
          
          {/* NEW LOGOUT BUTTON */}
          <button 
            onClick={handleLogout}
            className="text-gray-400 hover:text-red-400 px-2 py-2 text-sm font-medium transition-colors"
            title="Log Out"
          >
            Logout
          </button>
        </div>
      </nav>

      <div className="flex flex-1 max-w-7xl w-full mx-auto overflow-hidden">
        {/* ... (Your existing Sidebar and Main Feed remain exactly the same) ... */}
        <aside className="w-64 border-r border-gray-800 p-6 overflow-y-auto hidden md:block">
          <h3 className="text-xs font-bold text-gray-500 uppercase tracking-wider mb-4">Your Subscriptions</h3>
          <ul className="space-y-2">
            <li>
              <button 
                onClick={() => setActiveFeedId(null)}
                className={`w-full text-left px-3 py-2 rounded-lg text-sm font-medium transition-colors ${
                  activeFeedId === null ? 'bg-emerald-900/40 text-emerald-400' : 'text-gray-400 hover:bg-gray-800 hover:text-gray-200'
                }`}
              >
                All Articles
              </button>
            </li>
            {feeds.map(feed => (
              <li key={feed.id} className="group relative flex items-center">
                <button 
                  onClick={() => { setActiveFeedId(feed.id); setPage(1); }}
                  className={`w-full text-left px-3 py-2 rounded-lg text-sm font-medium transition-colors truncate pr-8 ${
                    activeFeedId === feed.id ? 'bg-emerald-900/40 text-emerald-400' : 'text-gray-400 hover:bg-gray-800 hover:text-gray-200'
                  }`}
                >
                  {feed.name}
                </button>
                
                {/* THE UNFOLLOW BUTTON (Hidden until you hover over the feed) */}
                <button
                  onClick={(e) => handleUnfollow(feed.id, e)}
                  className="absolute right-2 opacity-0 group-hover:opacity-100 text-gray-500 hover:text-red-400 transition-opacity p-1 text-xs font-bold"
                  title="Unfollow"
                >
                  ✕
                </button>
              </li>
            ))}
          </ul>
        </aside>

        <main className="flex-1 p-6 lg:p-10 overflow-y-auto">
          {loading && <p className="text-center text-gray-400 animate-pulse mt-10">Fetching your personalized feed...</p>}
          {error && <p className="text-center text-red-400 bg-red-900/20 p-4 rounded-lg border border-red-800/50">{error}</p>}

          <div className="max-w-3xl space-y-6">
            {!loading && !error && displayedPosts.length === 0 && (
              <div className="text-center mt-20">
                <p className="text-gray-500 text-lg">No articles found.</p>
                {feeds.length === 0 && (
                  <button onClick={() => setIsModalOpen(true)} className="mt-4 text-emerald-500 hover:underline">
                    Add your first RSS feed
                  </button>
                )}
              </div>
            )}
            
            {displayedPosts.map((post) => (
              <article key={post.id} className="bg-gray-900 border border-gray-800 rounded-xl p-6 shadow-sm hover:shadow-md transition-shadow">
                <a href={post.url} target="_blank" rel="noopener noreferrer" className="hover:text-emerald-400 transition-colors">
                  <h2 className="text-xl font-semibold mb-2">{post.title}</h2>
                </a>
                
                {/* 1. Standard Description Preview */}
                <p className="text-gray-400 text-sm leading-relaxed my-3 line-clamp-3">
                  {post.description || "No description available."}
                </p>

                {/* 2. The AI Summary Section (Only renders if the backend generated one!) */}
                {post.summary && (
                  <div className="my-4">
                    <button 
                      onClick={() => toggleSummary(post.id)}
                      className={`flex items-center space-x-2 text-xs font-semibold px-4 py-2 rounded-full transition-all ${
                        expandedSummaries[post.id] 
                          ? 'bg-emerald-900/30 text-emerald-400 border border-emerald-800/50' 
                          : 'bg-indigo-900/20 text-indigo-300 border border-indigo-800/30 hover:bg-indigo-900/40'
                      }`}
                    >
                      <span>✨</span>
                      <span>{expandedSummaries[post.id] ? 'Hide AI Summary' : 'Read AI Summary'}</span>
                    </button>

                    {/* 3. The Expanded AI Text Box */}
                    {expandedSummaries[post.id] && (
                      <div className="mt-3 p-5 bg-gradient-to-br from-gray-800 to-gray-900 border border-emerald-900/30 rounded-xl shadow-inner animate-in fade-in slide-in-from-top-2 duration-300">
                        <div className="flex items-center space-x-2 mb-3">
                          <span className="text-emerald-500 text-xs font-bold tracking-widest uppercase">Generated by Gemini</span>
                        </div>
                        <p className="text-gray-300 text-sm leading-relaxed">
                          {post.summary}
                        </p>
                      </div>
                    )}
                  </div>
                )}
                
                <div className="text-xs text-gray-500 font-medium flex items-center justify-between uppercase tracking-wide">
                  <span>{new Date(post.published_at).toLocaleDateString()}</span>
                  <span className="bg-gray-800 px-2 py-1 rounded text-gray-400">
                    {feeds.find(f => f.id === post.feed_id)?.name || 'Unknown Source'}
                  </span>
                </div>
              </article>
            ))}
            {/* The Load More Button */}
            {!loading && !error && displayedPosts.length > 0 && hasMore && !activeFeedId && !isSearching && (
              <div className="flex justify-center pt-8 pb-12">
                <button 
                  onClick={loadMorePosts}
                  disabled={isLoadingMore}
                  className="bg-gray-800 hover:bg-gray-700 text-gray-300 border border-gray-700 px-8 py-3 rounded-full font-medium transition-all disabled:opacity-50 flex items-center space-x-2"
                >
                  {isLoadingMore ? (
                    <>
                      <div className="w-4 h-4 border-2 border-emerald-500 border-t-transparent rounded-full animate-spin"></div>
                      <span>Loading...</span>
                    </>
                  ) : (
                    <span>Load More Articles</span>
                  )}
                </button>
              </div>
            )}
            
            {!hasMore && displayedPosts.length > 0 && (
              <p className="text-center text-gray-600 py-8">You've reached the end of your feed.</p>
            )}
          </div>
        </main>
      </div>

      <AddFeedModal 
        isOpen={isModalOpen} 
        onClose={() => setIsModalOpen(false)} 
        onFeedAdded={loadDashboardData} 
      />
    </div>
  );
}

export default App;