'use client'
import { useState, useEffect, useMemo,use } from 'react';
import axios from 'axios';
import { 
  Loader2, 
  AlertCircle, 
  CheckCircle2, 
  ListTodo, 
  AlertTriangle, 
  Calendar, 
  Tag,
  ChevronDown,
  ChevronUp,
  Plus,
  Check,
  MessageSquare
} from 'lucide-react';
import { format, formatDistanceToNow, isBefore, isToday } from 'date-fns';
import CreateTaskModal from '../../../../components/Modals/CreateTaskModal'; // Assuming this path

// Helper to format deadline with colors
function DeadlineDisplay({ deadline, isCompleted }) {
  const d = new Date(deadline);
  const now = new Date();
  
  let color = 'text-gray-500';
  let text = `Due ${formatDistanceToNow(d, { addSuffix: true })}`;
  
  if (!isCompleted) {
    if (isBefore(d, now) && !isToday(d)) {
      color = 'text-red-600 font-medium';
      text = `Overdue ${formatDistanceToNow(d, { addSuffix: true })}`;
    } else if (isToday(d)) {
      color = 'text-amber-600 font-medium';
      text = 'Due today';
    }
  } else {
    text = `Completed ${format(d, 'MMM d, yyyy')}`;
  }

  return (
    <div className={`flex items-center gap-1.5 ${color}`}>
      <Calendar size={14} />
      <span>{text}</span>
    </div>
  );
}

// Individual Task Item
function TaskItem({ task }) {
  const { title, description, deadline, status } = task;
  const isCompleted = status === 'completed';

  return (
    <div className="flex p-4 bg-white rounded-lg shadow-sm border border-gray-200 hover:shadow-md transition-shadow">
      {/* Custom Checkbox (Read-only for now) */}
      <div
        className={`
          flex-shrink-0 mt-1 w-6 h-6 rounded-md border-2 transition-all duration-200
          flex items-center justify-center
          ${isCompleted 
            ? 'bg-emerald-500 border-emerald-500' 
            : 'border-gray-400'
          }
        `}
      >
        {isCompleted && <Check size={16} className="text-white" />}
      </div>

      <div className="ml-4 flex-grow">
        <p className={`font-semibold text-gray-800 ${isCompleted ? 'line-through text-gray-500' : ''}`}>
          {title}
        </p>
        <p className={`text-sm text-gray-600 mt-1 ${isCompleted ? 'line-through text-gray-500' : ''}`}>
          {description}
        </p>
        <div className="flex items-center gap-4 text-sm text-gray-500 mt-2">
          <DeadlineDisplay deadline={deadline} isCompleted={isCompleted} />
        </div>
      </div>
    </div>
  );
}

export default function GroupTasksPage({ params }: { params: { groupId: string } }) {
  const [isOpen, setIsOpen] = useState(false);
  const [tasks, setTasks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const { groupId } = use(params);

  const getTasks = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await axios.get(`http://localhost:8080/groups/${groupId}/tasks`, {
        withCredentials: true
      });
      setTasks(response.data || []);
    } catch (err: any) {
      console.error(err);
      setError(err.response?.data?.message || "Failed to load tasks.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    getTasks();
  }, [groupId]);

  const { pendingTasks, completedTasks } = useMemo(() => {
    const pending = tasks
      .filter(t => t.status !== 'completed')
      .sort((a, b) => new Date(a.deadline).getTime() - new Date(b.deadline).getTime());
    const completed = tasks
      .filter(t => t.status === 'completed')
      .sort((a, b) => new Date(b.deadline).getTime() - new Date(a.deadline).getTime());
    return { pendingTasks: pending, completedTasks: completed };
  }, [tasks]);

  const renderContent = () => {
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

    if (tasks.length === 0) {
      return (
        <div className="text-center text-gray-500 mt-16 p-6 bg-gray-100 rounded-lg">
          <ListTodo size={48} className="mx-auto text-gray-400 mb-4" />
          <h3 className="text-xl font-semibold mb-2">No Tasks Yet</h3>
          <p>This group doesn't have any tasks. Create one to get started!</p>
        </div>
      );
    }

    return (
      <div className="space-y-8">
        {/* Pending Tasks Section */}
        <section>
          <div className="flex items-center gap-3 mb-4">
            <AlertTriangle className="text-amber-500" />
            <h2 className="text-xl font-semibold text-gray-700">Pending</h2>
            <span className="text-sm font-medium text-gray-500 bg-gray-200 px-2 py-0.5 rounded-full">
              {pendingTasks.length}
            </span>
          </div>
          <div className="space-y-3">
            {pendingTasks.length > 0 ? (
              pendingTasks.map((task) => <TaskItem key={task.id} task={task} />)
            ) : (
              <p className="text-gray-500 px-4 py-2">No pending tasks. Well done!</p>
            )}
          </div>
        </section>

        {/* Completed Tasks Section */}
        <section>
          <div className="flex items-center gap-3 mb-4">
            <CheckCircle2 className="text-emerald-500" />
            <h2 className="text-xl font-semibold text-gray-700">Completed</h2>
            <span className="text-sm font-medium text-gray-500 bg-gray-200 px-2 py-0.5 rounded-full">
              {completedTasks.length}
            </span>
          </div>
          <div className="space-y-3">
            {completedTasks.length > 0 ? (
              completedTasks.map((task) => <TaskItem key={task.id} task={task} />)
            ) : (
              <p className="text-gray-500 px-4 py-2">No completed tasks yet.</p>
            )}
          </div>
        </section>
      </div>
    );
  };

  return (
    <div className="max-w-[90%] mx-auto">
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold text-gray-800">Group Tasks</h1>
        <button
          className="bg-emerald-500 text-white px-4 py-2 rounded-lg shadow-md hover:bg-emerald-600 transition-all duration-300 flex items-center gap-2 font-medium"
          onClick={() => setIsOpen(true)}
        >
          <Plus size={18} />
          <span>Create Task</span>
        </button>
      </div>
      
      {renderContent()}

      {/* The Modal (using the placeholder) */}
      <CreateTaskModal 
        open={isOpen} 
        onClose={() => setIsOpen(false)} 
        groupId={groupId} 
        onTaskCreated={getTasks} // Add a prop to refresh tasks on creation
      />
    </div>
  )
}
