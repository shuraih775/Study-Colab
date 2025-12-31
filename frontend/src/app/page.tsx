'use client'

import { motion } from 'framer-motion'
import Link from 'next/link'
import AnimatedButton from '@/components/AnimatedButton'
import { useRouter } from 'next/navigation'
import GlowingButton from '@/components/GLowingButton'

export default function HomePage() {
  const router = useRouter();
  return (
    <div className="relative min-h-screen flex flex-col items-center justify-center overflow-hidden bg-gradient-to-br from-amber-400 via-teal-200 to-amber-400 text-center px-6">

      <motion.div
        className="absolute w-[40rem] h-[40rem] bg-indigo-300 rounded-full mix-blend-multiply filter blur-3xl opacity-20"
        animate={{
          x: [0, 100, -100, 0],
          y: [0, -50, 50, 0],
          scale: [1, 1.2, 1],
        }}
        transition={{
          duration: 15,
          repeat: Infinity,
          repeatType: 'mirror',
          ease: 'easeInOut',
        }}
      />
      <motion.div
        className="absolute w-[30rem] h-[30rem] bg-blue-300 rounded-full mix-blend-multiply filter blur-3xl opacity-20"
        animate={{
          x: [50, -100, 100, 50],
          y: [-50, 50, -50, -50],
          scale: [1.1, 1, 1.2],
        }}
        transition={{
          duration: 18,
          repeat: Infinity,
          repeatType: 'mirror',
          ease: 'easeInOut',
        }}
      />

      {/* Header */}
      <header className="relative w-full max-w-6xl mx-auto py-6 flex justify-between items-center z-10">
        <motion.h1
          className="text-2xl md:text-3xl font-bold text-indigo-700"
          initial={{ y: -30, opacity: 0 }}
          animate={{ y: 0, opacity: 1 }}
          transition={{ duration: 0.7 }}
        >
          📚 StudyBuddy
        </motion.h1>

        <motion.div
          className="space-x-4"
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.2 }}
        >
          
          <GlowingButton  onClick={()=>{router.push('/auth/login')}} className='bg-transparent border border-gray-100 hover:bg-transparent px-3 py-1' >
            <p className='font-small'>Log In</p>
          </GlowingButton>

          <GlowingButton onClick={()=>{router.push('/auth/register')}} className='px-3 py-1'>
                              <p className='font-small'>Register</p>
      
                      </GlowingButton>
          
        </motion.div>
      </header>

      {/* Main Section */}
      <main className="relative flex-1 flex flex-col items-center justify-center z-10">
        <motion.h2
          className="text-4xl md:text-5xl font-extrabold text-gray-900 mb-4"
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.3 }}
        >
          Collaborate. Learn. Succeed.
        </motion.h2>

        <motion.p
          className="text-lg text-gray-600 max-w-xl mb-8"
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.5 }}
        >
          Join study groups, share resources, manage tasks, and boost your learning journey 
          with peers across the world.
        </motion.p>

        <motion.div
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.7, delay: 0.7 }}
        >
            <AnimatedButton  onClick={()=>{router.push('/auth/register')}} className="bg-indigo-600 text-white cursor-pointer">
              Get Started
            </AnimatedButton>
        </motion.div>
      </main>

      {/* Footer */}
      <motion.footer
        className="relative text-sm text-gray-500 py-6 z-10"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 1.2 }}
      >
        © {new Date().getFullYear()} StudyBuddy. All rights reserved.
      </motion.footer>
    </div>
  )
}
