import React, { useEffect, useState } from 'react';
import axios from 'axios';
import { useRouter } from 'next/navigation';
import {
  Users,   
  Hash,   
  Plus,    
  AlertCircle, 
  Loader2  
} from 'lucide-react';
import CreateGroupsModal from '../../components/Modals/CreateGroupModal';



export default function Groups() {
  const [groups, setGroups] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const router = useRouter();
  const [showModal, setShowModal] = useState(false);
  

  useEffect(() => {
    const getGroups = async () => {
      setLoading(true);
      setError(null);
      try {
        const response = await axios.get(`http://localhost:8080/user/groups`, {
          withCredentials: true
        });
        setGroups(response.data.groups || []); 
      } catch (err) {
        console.error("Failed to fetch groups:", err);
        setError("Could not load your groups. Please try again later.");
      } finally {
        setLoading(false);
      }
    };

    getGroups();
  }, []);

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <Loader2 className="animate-spin text-emerald-500" size={48} />
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

  return (
    <div>
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold text-gray-800">My Groups</h1>
        <button className="bg-emerald-500 text-white px-4 py-2 rounded-lg shadow-md hover:bg-emerald-600 transition-all duration-300 flex items-center gap-2 font-medium cursor-pointer" onClick={()=>{setShowModal(true)}}>
          <Plus size={18} />
          <span>Create Group</span>
        </button>
      </div>

      {groups.length === 0 ? (
        <div className="text-center text-gray-500 mt-20 p-6 bg-gray-100 rounded-lg">
          <h3 className="text-xl font-semibold mb-2">No Groups Yet</h3>
          <p>You haven't joined or created any groups.</p>
          <p>Why not create one and get started?</p>
        </div>
      ) : (
        // 5. Groups Grid
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {groups.map((group) => (
            <GroupCard key={group.id} group={group} router={router} />
          ))}
        </div>
      )}
            {showModal && <CreateGroupsModal open={showModal} onClose={()=>{setShowModal(false)}}/> }
      
    </div>
  );
}


function GroupCard({ group, router }) {
  const initial = group.name ? group.name[0].toUpperCase() : '?';

  return (
    <div className="bg-white rounded-xl shadow-md overflow-hidden transition-all duration-300 hover:shadow-xl hover:-translate-y-1 flex flex-col">
      <div className="p-6 flex-grow">
        <div className="flex items-center gap-4 mb-4">
          {/* Avatar Placeholder */}
          <div className="flex-shrink-0 w-14 h-14 rounded-full bg-emerald-100 flex items-center justify-center border-2 border-emerald-200">
            <span className="text-2xl font-semibold text-emerald-700">{initial}</span>
          </div>
          {/* Group Name & Type */}
          <div>
            <h3 className="text-lg font-semibold text-gray-900 truncate" title={group.name}>
              {group.name}
            </h3>
            <div className="flex items-center text-sm text-gray-500 gap-1 mt-1">
              <Hash size={14} className="text-amber-500" />
              <span>{group.type}</span>
            </div>
          </div>
        </div>
        
        {/* Members Info */}
        <div className="flex items-center text-gray-600 mb-6">
          <Users size={16} className="text-gray-400 mr-2" />
          <span className="text-sm font-medium">{group.members} members</span>
        </div>
      </div>

      {/* Action Button */}
      <div className="p-6 pt-0">
        <button 
          onClick={() => router.push(`/groups/${group.id}/chat`)}
          className="w-full bg-emerald-50 text-emerald-700 py-2 px-4 rounded-lg font-semibold transition-colors duration-300 hover:bg-emerald-100 hover:text-emerald-800 focus:outline-none focus:ring-2 focus:ring-emerald-400 focus:ring-opacity-75"
        >
          Open Group
        </button>
      </div>
    </div>
  );
}