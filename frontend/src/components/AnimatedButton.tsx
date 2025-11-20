'use client'

import React from 'react'
import { motion } from 'framer-motion'

/**
 * Variants for the sliding overlay.
 * 'rest': The overlay is off-screen to the left.
 * 'hover': The overlay slides in to cover the button.
 */
const overlayVariants = {
  rest: {
    x: '-100%',
  },
  hover: {
    x: '0%',
  },
}

/**
 * Variants for the button's scale effect.
 */
const buttonVariants = {
  rest: {
    scale: 1,
  },
  hover: {
    scale: 1.05,
  },
}

/**
 * An animated button component that features a new background
 * sliding in from the left on hover.
 */
export default function AnimatedButton({ children, onClick, className = '' }) {
  return (
    <motion.button
      onClick={onClick}
      // Set initial and whileHover states to propagate variants to children
      initial="rest"
      whileHover="hover"
      whileTap={{ scale: 0.95 }}
      variants={buttonVariants}
      transition={{ // Spring transition for the scale effect
        type: 'spring',
        stiffness: 400,
        damping: 15,
      }}
      className={`
        relative overflow-hidden px-8 py-3 rounded-lg font-medium text-white
        bg-transparent border border-gray-100
        shadow-3xl focus:outline-none focus:ring-2 focus:ring-purple-400 focus:ring-opacity-75
        ${className}
      `}
    >
      {/* This div is the new background. 
        It sits on top of the button's base background (the indigo one) 
        and slides in on hover.
      */}
      <motion.div
        className="absolute inset-0 bg-gradient-to-r from-amber-300 via-amber-500 to-red-500"
        variants={overlayVariants}
        transition={{ // Tween transition for a smooth slide
          type: 'tween',
          duration: 0.4,
          ease: 'easeInOut',
        }}
      />

      {/* The span ensures the text stays above all backgrounds */}
      <span className="relative z-10">{children}</span>
    </motion.button>
  )
}

