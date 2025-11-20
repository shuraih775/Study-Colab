'use client'

import { motion } from 'framer-motion'

export default 
function GlowingButton({
  children,
  onClick,
  // Default glow color is now a soft indigo to match the button
  glowColor = 'rgba(165, 180, 252, 1)', // indigo-300
  className = '',
}) {
  // We dynamically create shadow strings with opacity from the base glowColor
  const subtleGlow = glowColor.replace('1)', '0.4)')
  const subtleTextGlow = glowColor.replace('1)', '0.5)')
  const hoverGlow = glowColor.replace('1)', '0.7)')

  return (
    <motion.button
      onClick={onClick}
      // "animate" is the resting state
      animate={{
        scale: 1,
        boxShadow: `0 0 10px ${subtleGlow}`,
        textShadow: `0 0 8px ${subtleTextGlow}`,
      }}
      // "whileHover" is the new state on hover
      whileHover={{
        scale: 1.05,
        boxShadow: `0 0 20px ${glowColor}, 0 0 35px ${hoverGlow}`,
        textShadow: `0 0 12px ${glowColor}`,
      }}
      // "whileTap" for the press effect
      whileTap={{ scale: 0.95 }}
      // A single spring transition handles all animations smoothly
      transition={{
        type: 'spring',
        stiffness: 300,
        damping: 10,
      }}
      className={`
        relative rounded-lg  text-white
        bg-indigo-600 
        focus:outline-none
        ${className}
      `}
    >
      {children}
    </motion.button>
  )
}