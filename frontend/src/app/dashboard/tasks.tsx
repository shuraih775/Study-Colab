import React, { useEffect, useState, useMemo } from 'react';
import axios from 'axios';
import { format, formatDistanceToNow, isBefore, isToday, addDays } from 'date-fns';
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
  Circle,
  Check
} from 'lucide-react';




/**
 * Main Tasks Page Component
 */
export default function Tasks() {
  const [allTasks, setAllTasks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const getTasks = async () => {
      setLoading(true);
      setError(null);
      try {
        const response = await axios.get(`http://localhost:8080/user/tasks`, {
          withCredentials: true
        });
        setAllTasks(response.data.tasks || []);
      } catch (err) {
        console.error("Failed to fetch tasks:", err);
        setError("Could not load tasks. Displaying mock data.");
      
      } finally {
        setLoading(false);
      }
    };
    getTasks();
  }, []);

  const handleToggleTask = (taskId) => {
    console.log(`(Mock API) Toggling task: ${taskId}`);
    setAllTasks(currentTasks =>
      currentTasks.map(task =>
        task.id === taskId
          ? { ...task, status: task.status === 'completed' ? 'pending' : 'completed' }
          : task
      )
    );
  };

  // Categorize tasks using useMemo for performance
  const { attentionTasks, upcomingTasks, completedTasks } = useMemo(() => {
    const now = new Date();
    const attention = [];
    const upcoming = [];
    const completed = [];

    allTasks.forEach(task => {
      if (task.status === 'completed') {
        completed.push(task);
        return;
      }

      const deadline = new Date(task.deadline);
      if (isBefore(deadline, now) || isToday(deadline)) {
        attention.push(task);
      } else {
        upcoming.push(task);
      }
    });

    // Sort by deadline
    const sortByDeadline = (a, b) => new Date(a.deadline) - new Date(b.deadline);
    attention.sort(sortByDeadline);
    upcoming.sort(sortByDeadline);
    completed.sort(sortByDeadline).reverse(); // Show most recently completed

    return { attentionTasks: attention, upcomingTasks: upcoming, completedTasks: completed };
  }, [allTasks]);

  // 1. Loading State
  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <Loader2 className="animate-spin text-emerald-500" size={48} />
      </div>
    );
  }

  // 2. Main Content
  return (
    <div>
      {/* Header */}
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold text-gray-800">My Tasks</h1>
        
      </div>

      {/* Optional Error Message */}
      {error && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded-lg flex items-center gap-3 mb-6">
          <AlertCircle size={20} />
          <span className="font-medium">{error}</span>
        </div>
      )}

      {/* 3. Empty State (if no tasks at all) */}
      {allTasks.length === 0 && !loading && (
        <div className="text-center text-gray-500 mt-20 p-6 bg-gray-100 rounded-lg">
          <CheckCircle2 size={48} className="mx-auto text-emerald-500 mb-4" />
          <h3 className="text-xl font-semibold mb-2">All Clear!</h3>
          <p>You have no tasks. Time for a coffee? ☕️</p>
        </div>
      )}

      {/* 4. Task Sections */}
      <div className="space-y-8">
        <TaskSection
          title="Needs Attention"
          tasks={attentionTasks}
          icon={<AlertTriangle className="text-amber-500" />}
          onToggleTask={handleToggleTask}
          defaultOpen={true}
        />
        <TaskSection
          title="Upcoming"
          tasks={upcomingTasks}
          icon={<ListTodo className="text-blue-500" />}
          onToggleTask={handleToggleTask}
          defaultOpen={true}
        />
        <TaskSection
          title="Completed"
          tasks={completedTasks}
          icon={<CheckCircle2 className="text-emerald-500" />}
          onToggleTask={handleToggleTask}
          defaultOpen={false} // Start with completed tasks collapsed
        />
      </div>
    </div>
  );
}

/**
 * Collapsible Task Section Component
 */
function TaskSection({ title, tasks, icon, onToggleTask, defaultOpen = true }) {
  const [isOpen, setIsOpen] = useState(defaultOpen);

  return (
    <section>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full flex justify-between items-center p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
      >
        <div className="flex items-center gap-3">
          {icon}
          <h2 className="text-xl font-semibold text-gray-700">{title}</h2>
          <span className="text-sm font-medium text-gray-500 bg-gray-200 px-2 py-0.5 rounded-full">
            {tasks.length}
          </span>
        </div>
        {isOpen ? <ChevronUp className="text-gray-500" /> : <ChevronDown className="text-gray-500" />}
      </button>

      {isOpen && (
        <div className="mt-4 space-y-3">
          {tasks.length === 0 ? (
            <p className="text-gray-500 px-4 py-2">No tasks in this section.</p>
          ) : (
            tasks.map(task => (
              <TaskItem key={task.id} task={task} onToggle={onToggleTask} />
            ))
          )}
        </div>
      )}
    </section>
  );
}

/**
 * Individual Task Item Component
 */
function TaskItem({ task, onToggle }) {
  const { id, title, deadline, status, groupName } = task;
  const isCompleted = status === 'completed';

  return (
    <div className="flex items-center p-4 bg-white rounded-lg shadow-sm border border-gray-200 hover:shadow-md transition-shadow">
      {/* Custom Checkbox */}
      <button
        onClick={() => onToggle(id)}
        className={`
          flex-shrink-0 w-6 h-6 rounded-md border-2 transition-all duration-200
          flex items-center justify-center
          ${isCompleted 
            ? 'bg-emerald-500 border-emerald-500' 
            : 'border-gray-400 hover:border-emerald-500'
          }
        `}
      >
        {isCompleted && <Check size={16} className="text-white" />}
      </button>

      {/* Task Details */}
      <div className="ml-4 flex-grow">
        <p className={`
          font-medium text-gray-800
          ${isCompleted ? 'line-through text-gray-500' : ''}
        `}>
          {title}
        </p>
        <div className="flex items-center gap-4 text-sm text-gray-500 mt-1">
          {/* Deadline */}
          <DeadlineDisplay deadline={deadline} isCompleted={isCompleted} />
          {/* Group Name */}
          {groupName && (
            <div className="flex items-center gap-1">
              <Tag size={14} />
              <span>{groupName}</span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

/**
 * Helper to display the formatted deadline with correct color
 */
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
    text = `Completed ${formatDistanceToNow(d, { addSuffix: true })}`;
  }

  return (
    <div className={`flex items-center gap-1 ${color}`}>
      <Calendar size={14} />
      <span>{text}</span>
    </div>
  );
}
