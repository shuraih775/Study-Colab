'use client'
import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { 
  Search, 
  Users, 
  Lock, 
  Star, 
  Clock, 
  CaseSensitive, 
  ChevronDown, 
  Loader2, 
  SearchX,
  Plus,
  LogIn,
  Check
} from 'lucide-react';
import axios from 'axios';
import { useRouter } from 'next/navigation';
import CreateGroupsModal from '../../components/Modals/CreateGroupModal';


export default function DiscoverGroups() {
  const [allGroups, setAllGroups] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [showModal, setShowModal] = useState(false);
  const [filters, setFilters] = useState({
    type: 'all',
    sortBy: 'popularity'
  });


  const getGroups = useCallback(
      async()=>{
      await axios.get('http://localhost:8080/groups/')
      .then(res => setAllGroups(res.data))
      .catch(err => console.error(err))
      .finally(()=>{setLoading(false)});

    },[]);

  useEffect(()=>{
    
    getGroups();
    
  },[])

  const handleFilterChange = (e) => {
    const { name, value } = e.target;
    setFilters(prev => ({ ...prev, [name]: value }));
  };

  // Memoized filtering logic
  const filteredGroups = useMemo(() => {
    let groups = [...allGroups];

    // 1. Filter by Search Query
    if (searchQuery) {
      const lowerQuery = searchQuery.toLowerCase();
      groups = groups.filter(group => 
        group.name.toLowerCase().includes(lowerQuery) || 
        group.description.toLowerCase().includes(lowerQuery)
      );
    }

    // 2. Filter by Type
    if (filters.type !== 'all') {
      groups = groups.filter(group => group.type === filters.type);
    }

    // 3. Sort
    switch (filters.sortBy) {
      case 'popularity':
        groups.sort((a, b) => b.members - a.members);
        break;
      case 'newest':
        groups.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
        break;
      case 'alphabetical':
        groups.sort((a, b) => a.name.localeCompare(b.name));
        break;
    }

    return groups;
  }, [allGroups, searchQuery, filters]);

  return (
    <div className="p-4 md:p-8 max-w-7xl mx-auto">
      <div className='flex flex-row justify-between mb-4 mt-4'>

      <h1 className="text-3xl font-bold text-gray-800">Discover Groups</h1>
      <button className='bg-emerald-300 p-4 rounded-lg hover:bg-emerald-400 hover:-translate-y-2 transition-all duration-300'onClick={()=>{setShowModal(true)}}> Create Group</button>

      </div>


      
      {/* Filter Bar */}
      <FilterBar 
        searchQuery={searchQuery} 
        onSearchChange={(e) => setSearchQuery(e.target.value)}
        filters={filters}
        onFilterChange={handleFilterChange}
      />

      
      {/* Content Area */}
      <div className="mt-8">
        {loading ? (
          <LoadingSpinner />
        ) : filteredGroups.length === 0 ? (
          <EmptyState query={searchQuery} />
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {filteredGroups.map(group => (
              <GroupCard 
                key={group.id} 
                group={group}
                isMember={false} // to be implemented later
                isPending={false} // to be implemented later
              />
            ))}
          </div>
        )}
      </div>
      
      {showModal && <CreateGroupsModal open={showModal} onClose={()=>{setShowModal(false)}}/> }
    </div>
  );
}

/**
 * Search and Filter Bar Component
 */
function FilterBar({ searchQuery, onSearchChange, filters, onFilterChange }) {
  return (
    <div className="bg-white p-4 rounded-xl shadow-lg border border-gray-200 flex flex-col md:flex-row gap-4">
      {/* Search Input */}
      <div className="relative flex-grow">
        <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
          <Search className="text-gray-400" size={20} />
        </div>
        <input
          type="text"
          placeholder="Search groups by name or description..."
          className="w-full pl-10 p-3 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-amber-400"
          value={searchQuery}
          onChange={onSearchChange}
        />
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        {/* Type Filter */}
        <div className="relative">
          <select 
            name="type"
            value={filters.type}
            onChange={onFilterChange}
            className="w-full sm:w-48 appearance-none bg-white p-3 pr-10 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-amber-400"
          >
            <option value="all">All Types</option>
            <option value="public">Public</option>
            <option value="private">Private</option>
          </select>
          <ChevronDown className="absolute right-3 top-3.5 text-gray-400 pointer-events-none" size={20} />
        </div>

        {/* Sort Filter */}
        <div className="relative">
          <select 
            name="sortBy"
            value={filters.sortBy}
            onChange={onFilterChange}
            className="w-full sm:w-48 appearance-none bg-white p-3 pr-10 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-amber-400"
          >
            <option value="popularity">Sort by Popularity</option>
            <option value="newest">Sort by Newest</option>
            <option value="alphabetical">Sort Alphabetically</option>
          </select>
          <ChevronDown className="absolute right-3 top-3.5 text-gray-400 pointer-events-none" size={20} />
        </div>
      </div>
    </div>
  );
}

/**
 * Group Card Component
 */
function GroupCard({ group, isMember, isPending }) {
  const router = useRouter();
  const {id, name, description, type, members } = group;
  
  const handleJoin = () => {
    router.push(`/groups/${id}/chat`)
  }

  // Determine button state
  let ButtonComponent;
  if (isMember) {
    ButtonComponent = (
      <button 
        disabled
        className="w-full flex items-center justify-center gap-2 bg-amber-100 text-amber-700 py-2 px-4 rounded-lg font-semibold cursor-default"
      >
        <Check size={18} />
        Joined
      </button>
    );
  } else if (isPending) {
    ButtonComponent = (
      <button 
        disabled 
        className="w-full flex items-center justify-center gap-2 bg-gray-200 text-gray-500 py-2 px-4 rounded-lg font-semibold cursor-wait"
      >
        <Clock size={18} />
        Pending
      </button>
    );
  } else if (type === 'private') {
    ButtonComponent = (
      <button 
        onClick={handleJoin}
        className="w-full flex items-center justify-center gap-2 bg-amber-50 text-amber-700 py-2 px-4 rounded-lg font-semibold transition-colors duration-300 hover:bg-amber-100 hover:text-amber-800"
      >
        <Plus size={18} />
        Request to Join
      </button>
    );
  } else {
    ButtonComponent = (
      <button 
        onClick={handleJoin}
        className="w-full flex items-center justify-center gap-2 bg-emerald-500 text-white py-2 px-4 rounded-lg font-semibold transition-colors duration-300 hover:bg-emerald-600"
      >
        <LogIn size={18} />
        Join
      </button>
    );
  }

  return (
    <div className="bg-white rounded-xl shadow-lg border border-gray-200 overflow-hidden flex flex-col transition-all duration-300 hover:shadow-xl hover:-translate-y-1">
      {/* Card Header */}
      <div className="p-5">
        <div className="flex items-center gap-4 mb-4">
          {/* Avatar Placeholder */}
          <div className={`
            flex-shrink-0 w-14 h-14 rounded-full flex items-center justify-center border-2
            ${type === 'private' 
              ? 'bg-amber-100 border-amber-200' 
              : 'bg-emerald-100 border-emerald-200'
            }
          `}>
            <span className={`text-2xl font-semibold 
              ${type === 'private' 
                ? 'text-amber-700' 
                : 'text-emerald-700'
              }
            `}>
              {name[0].toUpperCase()}
            </span>
          </div>
          {/* Group Name & Type */}
          <div>
            <h3 className="text-lg font-semibold text-gray-900 truncate" title={name}>
              {name}
            </h3>
            <span className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-medium mt-1 ${
              type === 'private' 
                ? 'bg-amber-100 text-amber-700' 
                : 'bg-emerald-100 text-emerald-700'
            }`}>
              {type === 'private' ? <Lock size={12} /> : <Users size={12} />}
              {type.charAt(0).toUpperCase() + type.slice(1)}
            </span>
          </div>
        </div>

        {/* Description */}
        <p className="text-sm text-gray-600 mb-4 h-10 line-clamp-2">
          {description}
        </p>

        {/* Member Count */}
        <div className="flex items-center text-gray-500">
          <Users size={16} className="text-gray-400 mr-2" />
          <span className="text-sm font-medium">{members.toLocaleString()} members</span>
        </div>
      </div>

      {/* Action Button */}
      <div className="p-5 pt-0 mt-auto">
        {ButtonComponent}
      </div>
    </div>
  );
}

/**
 * Loading Spinner Component
 */
function LoadingSpinner() {
  return (
    <div className="flex justify-center items-center h-64">
      <Loader2 className="animate-spin text-amber-600" size={48} />
    </div>
  );
}

/**
 * Empty State Component
 */
function EmptyState({ query }) {
  return (
    <div className="text-center text-gray-500 mt-20 p-6 bg-gray-50 rounded-lg">
      <SearchX size={48} className="mx-auto text-gray-400 mb-4" />
      <h3 className="text-xl font-semibold mb-2">No Groups Found</h3>
      <p>
        {query 
          ? `We couldn't find any groups matching "${query}".`
          : "Try adjusting your filters to find more groups."
        }
      </p>
    </div>
  );
}
