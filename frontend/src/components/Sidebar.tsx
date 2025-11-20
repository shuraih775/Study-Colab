import Link from 'next/link'

const Sidebar = () => {
  return (
    <aside className="w-64 bg-gray-800 text-white flex flex-col p-4">
      <h2 className="text-xl font-bold mb-6">StudyBuddy</h2>
      <nav className="flex flex-col gap-4">
        <Link href="/dashboard" className="hover:text-indigo-400">Dashboard</Link>
        <Link href="/groups/demo-group" className="hover:text-indigo-400">Groups</Link>
      </nav>
    </aside>
  )
}

export default Sidebar
