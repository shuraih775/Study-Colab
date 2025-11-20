'use client'

import React, { useState } from 'react';
import {
  LayoutDashboard,
  BookOpen,
  ClipboardList,
  Users,
  Video,
  Settings,
  Flag,
  HelpCircle,
  LogOut,
  Plus,
  BarChart3 // Using this as a placeholder for the logo
} from 'lucide-react';
import { useRouter } from 'next/navigation';
import axios from 'axios';

// Assuming these components are in the same directory
import Overview from './Overview'
import Groups from './groups';
import Tasks from './tasks';


/**
 * Main App component
 * This component renders the layout, including the sidebar and a main content area.
 */
export default function App({}) { // Renamed to App to match export
  // State to track the active navigation item
  const [activeItem, setActiveItem] = useState('Overview');
  const router = useRouter();

  // --- Logout Function ---
  /**
   * Clears all document cookies and redirects to the login page.
   */
  const handleLogout = async () => {
    console.log('Logging out...');
    
    
    await axios.post('http://localhost:8080/auth/logout',{},{
      withCredentials:true
    }).then(()=>{
        console.log('success')
    }).catch((error)=>{
      console.log(error);

    })
    
    
    // router.push('/auth/login');
    window.location.href = '/auth/login';
  };
  // -----------------------

  // Navigation items for the main menu
  const navItems = [
    { name: 'My Groups', icon: BookOpen, alert: true },
    { name: 'Tasks', icon: ClipboardList, alert: true },
    // { name: 'Members', icon: Users, alert: true },
    // { name: 'Meeting', icon: Video },
    { name: 'Setting', icon: Settings },
  ];

  // Navigation items for the bottom utility menu
  // Updated onClick handlers
  const bottomItems = [
    // { name: 'Report', icon: Flag, onClick: () => setActiveItem('Report') },
    // { name: 'Help', icon: HelpCircle, onClick: () => setActiveItem('Help') },
    { name: 'Log out', icon: LogOut, onClick: handleLogout }, // Use the logout function
  ];

  // Helper function to render the correct component based on the active item
  const renderContent = () => {
    switch (activeItem) {
      case 'Overview':
        return <Overview />;
      case 'My Groups': // Fixed typo from 'My groups'
        return <Groups />;
      case 'Tasks':
        return <Tasks />;
      case 'Setting':
        return <div><h1 className="text-3xl font-bold">Settings</h1></div>;
      // Added cases for Report and Help
      case 'Report':
        return <div><h1 className="text-3xl font-bold">Report</h1></div>;
      case 'Help':
        return <div><h1 className="text-3xl font-bold">Help</h1></div>;
      // 'Log out' doesn't need a case as it redirects
      default:
        return <Overview />;
    }
  };

  return (
    // Main container for the whole page layout
    <div className="flex h-screen bg-gray-50">
      
      {/* Sidebar Container */}
      <aside className="w-72 bg-white p-6 flex flex-col justify-between shadow-lg">
        
        {/* Top Section: Logo + Navigation */}
        <div>
          {/* Logo */}
          <div className="flex items-center gap-2 mb-10 px-2">
            <BarChart3 size={32} className="text-purple-600" />
            <span className="text-2xl font-bold text-gray-800">Study Colab</span>
          </div>

          {/* Navigation Menu */}
          <nav className="flex flex-col gap-2">
            
            {/* Dashboard Item (Special Case) */}
            {/* This item has a unique style when active, including a '+' icon */}
            <button
              onClick={() => setActiveItem('Overview')}
              className={`
                flex items-center justify-between p-3 rounded-lg cursor-pointer
                transition-colors
                ${
                  activeItem === 'Overview'
                    ? 'bg-emerald-400 text-white' // Active state
                    : 'text-gray-600 hover:bg-gray-100' // Inactive state
                }
              `}
            >
              <div className="flex items-center gap-3">
                <LayoutDashboard size={20} />
                <span className="font-medium">Overview</span>
              </div>
              {/* Show Plus icon only when Overview is active */}
              {/* {activeItem === 'Overview' && <Plus size={16} />} */}
            </button>

            {/* Other Navigation Items */}
            {navItems.map((item) => (
              <SidebarItem
                key={item.name}
                icon={<item.icon size={20} />}
                text={item.name}
                active={activeItem === item.name}
                alert={item.alert}
                onClick={() => setActiveItem(item.name)}
              />
            ))}
          </nav>
        </div>

        {/* Bottom Section: Utility Links */}
        <div className="flex flex-col gap-2 border-t border-gray-200 pt-6">
          {/* Updated this map to use the item.onClick directly */}
          {bottomItems.map((item) => (
            <SidebarItem
              key={item.name}
              icon={<item.icon size={20} />}
              text={item.name}
              active={activeItem === item.name}
              onClick={item.onClick} // Use the onClick from the item object
            />
          ))}
        </div>
      </aside>

      {/* Main Content Area (Placeholder) */}
      <main className="flex-1 p-10 overflow-y-auto">
        {renderContent()}
      </main>
    </div>
  );
}

/**
 * Helper component for individual sidebar items.
 * This makes the code cleaner and easier to manage.
 */
function SidebarItem({ icon, text, active, alert, onClick }) {
  return (
    <button
      onClick={onClick}
      className={`
        flex items-center justify-between w-full p-3 rounded-lg cursor-pointer
        transition-colors
        ${
          active
            ? 'bg-emerald-400 text-white' // Active state
            : 'text-gray-600 hover:bg-gray-100' // Inactive state
        }
      `}
    >
      <div className="flex items-center gap-3">
        {icon}
        <span className="font-medium">{text}</span>
      </div>
      {/* Mimics the small circle icons from the image */}
      {alert && !active && (
        <div className="w-2 h-2 rounded-full bg-gray-300" />
      )}
    </button>
  );
}