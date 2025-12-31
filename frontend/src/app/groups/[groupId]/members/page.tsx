'use client'
import React, { useState, useEffect, useMemo, use, useCallback } from 'react';
import { 
  Search, 
  Loader2, 
  AlertCircle, 
  User, 
  ShieldCheck, 
  ArrowUpCircle, 
  ArrowDownCircle, 
  UserX,
  Shield
} from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';
import axios from 'axios';

const CURRENT_USER_ID = 'u1'; // 
const CURRENT_USER_ROLE = 'admin'; 

export default function GroupMembersPage({params}) {
  const [members, setMembers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [searchQuery, setSearchQuery] = useState('');
  const { groupId } = use(params); // Assuming 'use(params)' is your way of getting route params

  const getMembers = useCallback(async () => {
    setLoading(true);
    try {
      // Make sure to replace with your actual API host
      const res = await axios.get(`http://localhost:8080/groups/${groupId}/members`, { withCredentials: true });
      setMembers(res.data.members);
      setError(null);
    } catch (err) {
      console.error(err);
      setError("Failed to fetch group members.");
    } finally {
      setLoading(false);
    }
  }, [groupId]); 

  const updateUser = async (operation, userId) => {
    let payload = {};

    switch (operation) {
      case 'Promote':
        payload = { role: 'admin' };
        break;
      case 'Demote':
        payload = { role: 'member' };
        break;
      case 'Kick':
        payload = { status: 'kicked' };
        break;
      default:
        console.error("Unknown operation:", operation);
        return;
    }

    try {
      await axios.put(`http://localhost:8080/groups/${groupId}/members/${userId}`, payload, { withCredentials: true });
      await getMembers(); 
    } catch (err) {
      console.error(err);
      setError(`Failed to ${operation.toLowerCase()} member.`);
    }
  };

  useEffect(() => {
    if (groupId) { 
      getMembers();
    }
  }, [getMembers, groupId]); 

  const filteredMembers = useMemo(() => {
    if (!searchQuery) return members;
    
    return members.filter(member =>
      member.username && member.username.toLowerCase().includes(searchQuery.toLowerCase())
    );
  }, [members, searchQuery]);

 
  const renderContent = () => {
    if (loading) {
      return (
        <div className="flex justify-center items-center h-64">
          <Loader2 className="animate-spin text-purple-600" size={48} />
        </div>
      );
    }

    if (error) {
      return (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded-lg flex items-center gap-3">
          <AlertCircle size={20} />
          <span className="font-medium">{error}</span>
        </div>
      );
    }

    if (filteredMembers.length === 0) {
      return (
        <div className="text-center text-gray-500 mt-16 p-6">
          <Search size={48} className="mx-auto text-gray-400 mb-4" />
          <h3 className="text-xl font-semibold mb-2">No Members Found</h3>
          <p>
            {searchQuery 
              ? "We couldn't find any members matching your search."
              : "This group doesn't have any members yet."
            }
          </p>
        </div>
      );
    }

    return (
      <ul className="divide-y divide-gray-200">
        {filteredMembers.map(member => (
          <MemberCard
            key={member.id}
            member={member}
            isCurrentUser={member.id === CURRENT_USER_ID} 
            currentUserIsAdmin={CURRENT_USER_ROLE === 'admin'} 
            onAction={updateUser}
          />
        ))}
      </ul>
    );
  };

  return (
    <div className="max-w-4xl mx-auto p-4 sm:p-6">
      <h1 className="text-3xl font-bold text-gray-800 mb-8">Group Members</h1>
      
      {/* Search Bar */}
      <div className="relative mb-6">
        <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
          <Search className="text-gray-400" size={20} />
        </div>
        <input
          type="text"
          placeholder="Search members by name..."
          className="w-full pl-12 p-3 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-purple-500"
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
        />
      </div>

      {/* Member List Container */}
      <div className="bg-white rounded-xl shadow-lg border border-gray-200 overflow-hidden">
        {renderContent()}
      </div>
    </div>
  );
}

function MemberCard({ member, isCurrentUser, currentUserIsAdmin, onAction }) {
  const { id, username, role, joined_at, avatarUrl, email } = member;
  const initial = username ? username.split(' ').map(n => n[0]).join('') : '?';

  return (
    <li className="p-4 sm:p-6 flex flex-col sm:flex-row justify-between items-start sm:items-center">
      {/* Left Side: User Info */}
      <div className="flex items-center gap-4">
        {/* Avatar */}
        <div className="flex-shrink-0 w-12 h-12 rounded-full overflow-hidden bg-gray-200">
          {avatarUrl ? (
            <img 
              src={avatarUrl} 
              alt={username} 
              className="w-full h-full object-cover" 
              onError={(e) => e.target.style.display = 'none'} // Hide if image fails to load
            />
          ) : (
            <div className="w-full h-full bg-purple-100 text-purple-600 flex items-center justify-center border-2 border-purple-200">
              <span className="text-lg font-semibold">{initial}</span>
            </div>
          )}
        </div>
        {/* Name, Role, Joined Date */}
        <div>
          <div className="font-semibold text-gray-800 flex items-center gap-2">
            {username || "Unknown User"}
            {isCurrentUser && (
              <span className="text-xs font-medium bg-purple-100 text-purple-700 px-2 py-0.5 rounded-full">
                You
              </span>
            )}
            {email && (
              <>
                <span className='text-stone-400'>•</span>
                <p className='text-stone-500 text-xs'>({email})</p>
              </>
            )}
          </div>

          <div className="flex items-center gap-3 text-sm text-gray-500 mt-1">
            <RoleBadge role={role} />
            {joined_at && (
              <>
                <span>•</span>
                <span>Joined {formatDistanceToNow(new Date(joined_at), { addSuffix: true })}</span>
              </>
            )}
          </div>
        </div>
      </div>

      {/* Right Side: Admin Controls */}
      {currentUserIsAdmin && !isCurrentUser && (
        <div className="flex gap-2 w-full sm:w-auto mt-4 sm:mt-0">
          {role === 'member' ? (
            <button
              onClick={() => onAction('Promote', id)}
              className="flex-1 sm:flex-none flex items-center justify-center gap-1.5 px-3 py-2 text-sm font-medium bg-emerald-100 text-emerald-700 rounded-lg hover:bg-emerald-200 transition-colors"
            >
              <ArrowUpCircle size={16} />
              Promote
            </button>
          ) : (
            <button
              onClick={() => onAction('Demote', id)}
              className="flex-1 sm:flex-none flex items-center justify-center gap-1.5 px-3 py-2 text-sm font-medium bg-amber-100 text-amber-700 rounded-lg hover:bg-amber-200 transition-colors"
            >
              <ArrowDownCircle size={16} />
              Demote
            </button>
          )}
          <button
            onClick={() => onAction('Kick', id)}
            className="flex-1 sm:flex-none flex items-center justify-center gap-1.5 px-3 py-2 text-sm font-medium bg-red-100 text-red-700 rounded-lg hover:bg-red-200 transition-colors"
          >
            <UserX size={16} />
            Kick
          </button>
        </div>
      )}
    </li>
  );
}


function RoleBadge({ role }) {
  if (role === 'admin') {
    return (
      <span className="inline-flex items-center gap-1.5 text-emerald-700">
        <ShieldCheck size={14} />
        <span className="font-medium">Admin</span>
      </span>
    );
  }
  return (
    <span className="inline-flex items-center gap-1.5 text-gray-600">
      <User size={14} />
      <span>Member</span>
    </span>
  );
}
