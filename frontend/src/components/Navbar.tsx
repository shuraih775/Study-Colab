const Topbar = () => {
  return (
    <header className="h-16 bg-white border-b flex items-center justify-between px-6">
      <h1 className="text-xl font-semibold">Welcome to StudyBuddy</h1>
      <div className="flex items-center gap-4">
        <button className="bg-indigo-500 text-white px-4 py-1 rounded-md hover:bg-indigo-600">Log out</button>
      </div>
    </header>
  )
}

export default Topbar
