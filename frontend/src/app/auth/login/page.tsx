'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import axios from 'axios'

export default function LoginPage() {
  const router = useRouter()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')

   const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      const res = await axios.post('http://localhost:8080/auth/login',
        { email, password },
        { withCredentials: true }
      );
     

      router.push('/dashboard')
      console.log(await cookieStore.getAll())
    } catch (err) {
      setError('login failed')
      console.error(err)
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br to-amber-400 from-emerald-400 text-white flex items-center justify-center px-4">
      <div className="w-full max-w-5xl grid md:grid-cols-2 bg-stone-200 rounded-2xl overflow-hidden shadow-2xl border border-stone-600">

        {/* Left Section - Login Form */}
        <div className="p-10 flex flex-col justify-center">
          <h2 className="text-3xl font-bold text-stone-800 text-center mb-8 tracking-wide">Login</h2>

          <button className="mb-6 flex items-center justify-center border border-stone-700 rounded-md py-2 hover:bg-stone-400 transition text-sm font-medium text-stone-800 hover:text-stone-100 cursor-pointer">
            <svg className="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 48 48">
              <path d="M44.5 20H24v8.5h11.9C34.8 33.4 30.2 37 24 37c-7.2 0-13-5.8-13-13s5.8-13 13-13c3.1 0 6 .9 8.3 2.5l6.2-6.2C34.2 4.8 29.3 3 24 3 12.4 3 3 12.4 3 24s9.4 21 21 21c10.8 0 19.8-7.8 21-18v-7z" />
            </svg>
            Sign in with Google
          </button>

          <div className="flex items-center mb-6">
            <hr className="flex-grow border-stone-700" />
            <span className="mx-3 text-stone-500 text-sm">or</span>
            <hr className="flex-grow border-stone-700" />
          </div>

          {error && <p className="text-red-500 text-sm mb-4 text-center">{error}</p>}

          <form onSubmit={handleLogin} className="space-y-4">
            <input
              type="email"
              placeholder="Email ID / Username"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full px-4 py-2 border border-stone-700 text-stone-800 rounded-md placeholder-stone-500 focus:ring-2 focus:ring-amber-500 outline-none"
              required
            />

            <input
              type="password"
              placeholder="Password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full px-4 py-2 border border-stone-700 text-stone-800 rounded-md placeholder-stone-500 focus:ring-2 focus:ring-amber-500 outline-none"
              required
            />

            <div className="flex items-center justify-between text-sm text-stone-500">
              <label className="flex items-center">
                <input type="checkbox" className="mr-2 accent-amber-500" />
                Remember me
              </label>
              <a href="#" className="text-stone-800 hover:underline">Forgot password?</a>
            </div>

            <button
              type="submit"
              className="w-full bg-amber-600 hover:bg-amber-700 transition py-2 rounded-md font-semibold shadow-md"
            >
              Login
            </button>
          </form>
        </div>

        {/* Right Section - Register Redirect */}
        <div className="bg-gradient-to-br from-emerald-300 to-emerald-600 flex flex-col items-center justify-center p-10 text-center">
          <h3 className="text-3xl font-bold mb-4">New Here?</h3>
          <p className="text-stone-200 mb-6">Join our community and start<br />collaborative learning today</p>
          <a href="/auth/register" className="border border-white px-6 py-2 rounded-full text-stone-200 hover:bg-emerald-600 hover:text-white transition">
            Register
          </a>
        </div>
      </div>
    </div>
  )
}
