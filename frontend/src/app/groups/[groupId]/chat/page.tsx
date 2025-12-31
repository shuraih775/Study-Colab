'use client'

import { useState, useEffect, useRef, use } from 'react'
import { Send, Paperclip, X, FileText, Image as ImageIcon } from 'lucide-react'
import { format } from 'date-fns'
import axios from 'axios'

const FAKE_CURRENT_USER = { 
  id: 'currentUser', 
  name: 'You'        
}

interface User {
  id: string;
  name: string;
}

interface Attachment {
  id: string;
  file_url: string;
  file_type: string;
  file_size: number;
}

interface Message {
  id: string;
  user: User;
  text: string;
  attachments: Attachment[]; 
  timestamp: string;
}

// --- COMPONENT ---

export default function GroupChatPage({ params }: { params: { groupId: string } }) {
  const { groupId } = use(params)
  
  // State
  const [messages, setMessages] = useState<Message[]>([])
  const [newMessage, setNewMessage] = useState('')
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [isUploading, setIsUploading] = useState(false)

  // Refs
  const ws = useRef<WebSocket | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    ws.current = new WebSocket("ws://localhost:8080/chat")

    ws.current.onopen = () => {
      console.log('WebSocket connected')
      const joinMsg = { type: 'join', room: groupId }
      ws.current?.send(JSON.stringify(joinMsg))
    }

    ws.current.onmessage = (event) => {
      const msg = JSON.parse(event.data) as Message
      setMessages((prev) => [...prev, msg])
      scrollToBottom()
    }

    ws.current.onclose = () => console.log('WebSocket disconnected')

    return () => {
      ws.current?.close()
    }
  }, [groupId])

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }



const uploadFileToMinIO = async (file: File) => {
  try {
    const res = await axios.post(`http://localhost:8080/materials/presigned-url`, 
      {
        file_name: file.name,
        file_type: file.type
      },
      {
        withCredentials: true,
        headers: {
          'Content-Type': 'application/json'
        }
      }
    );

    const { upload_url, file_id } = res.data;

await axios.put(upload_url, file, {
  headers: { 'Content-Type': file.type }
});

return { file_id };


  } catch (error) {
    console.error("Upload error:", error);
    alert("File upload failed");
    return null;
  }
}

  const handleSend = async (e: React.FormEvent) => {
    e.preventDefault()
    if ((newMessage.trim() === '' && !selectedFile) || !ws.current) return

    let attachmentsPayload: any[] = []

    if (selectedFile) {
      setIsUploading(true)
      const result = await uploadFileToMinIO(selectedFile)
      setIsUploading(false)

      if (result) {
        attachmentsPayload.push({
        file_id: result.file_id,
        file_name: selectedFile.name,
        file_type: selectedFile.type,
        file_size: selectedFile.size
      })

      } else {
        return 
      }
    }

    const messagePayload = {
  type: 'broadcast',
  room: groupId,
  text: newMessage,
  attachments: attachmentsPayload
}

    
    ws.current.send(JSON.stringify(messagePayload))
    
    setNewMessage('')
    setSelectedFile(null)
  }

  return (
    <div className="flex flex-col h-[calc(100vh-8rem)] max-w-[90%] mx-auto">
      <h1 className="text-3xl font-bold text-gray-800 mb-6">Group Chat</h1>
      
      {/* --- MESSAGES AREA --- */}
      <div className="flex-1 bg-white border border-gray-200 rounded-xl shadow-sm p-6 overflow-y-auto space-y-6">
        {messages.map((msg) => {
          const isCurrentUser = msg.user.id === FAKE_CURRENT_USER.id
          return (
            <div key={msg.id} className={`flex gap-3 ${isCurrentUser ? 'justify-end' : 'justify-start'}`}>
              
              {/* Avatar */}
              {!isCurrentUser && (
                <div className="flex-shrink-0 w-10 h-10 rounded-full bg-gray-200 flex items-center justify-center">
                  <span className="font-semibold text-gray-600">{msg.user.name[0]}</span>
                </div>
              )}

              {/* Message Bubble */}
              <div className={`flex flex-col ${isCurrentUser ? 'items-end' : 'items-start'} max-w-[75%]`}>
                <div 
                  className={`
                    p-3 rounded-2xl 
                    ${isCurrentUser 
                      ? 'bg-amber-400 text-white rounded-br-none' 
                      : 'bg-gray-100 text-gray-800 rounded-bl-none'
                    }
                  `}
                >
                  {!isCurrentUser && (
                    <div className="font-semibold text-sm mb-1 text-emerald-500">{msg.user.name}</div>
                  )}

                  {/* --- RENDER ATTACHMENTS --- */}
                  {msg.attachments && msg.attachments.length > 0 && (
                    <div className="space-y-2 mb-2">
                      {msg.attachments.map((att) => (
                        <div key={att.id}>
                          {att.file_type.startsWith('image/') ? (
                            // IMAGE PREVIEW
                            <img 
                              src={att.file_id} 
                              alt="attachment" 
                              className="rounded-lg max-h-60 object-cover border border-white/20"
                            />
                          ) : (
                            <a 
                              href={att.file_url} 
                              target="_blank"
                              rel="noopener noreferrer"
                              className="flex items-center gap-2 bg-black/10 p-2 rounded hover:bg-black/20 transition"
                            >
                              <FileText size={16} />
                              <span className="text-sm underline truncate max-w-[200px]">Download File</span>
                            </a>
                          )}
                        </div>
                      ))}
                    </div>
                  )}

                  {/* Text Content */}
                  {msg.text && <p>{msg.text}</p>}
                </div>
                
                <span className="text-xs text-gray-400 mt-1.5 px-1">
                  {format(new Date(msg.timestamp), 'h:mm a')}
                </span>
              </div>
            </div>
          )
        })}
        <div ref={messagesEndRef} />
      </div>
      
      {/* --- INPUT AREA --- */}
      <div className="mt-6">
        {/* File Preview Badge */}
        {selectedFile && (
            <div className="mb-2 flex items-center gap-2 bg-gray-100 w-fit px-3 py-1 rounded-full border border-gray-300">
                <span className="text-xs font-medium text-gray-600 truncate max-w-[200px]">
                    {selectedFile.name}
                </span>
                <button 
                    onClick={() => {
                        setSelectedFile(null)
                        if(fileInputRef.current) fileInputRef.current.value = ''
                    }}
                    className="text-gray-400 hover:text-red-500"
                >
                    <X size={14} />
                </button>
            </div>
        )}

        <form onSubmit={handleSend} className="flex gap-3">
            {/* Hidden File Input */}
            <input 
                type="file" 
                ref={fileInputRef}
                className="hidden"
                onChange={(e) => {
                    if (e.target.files && e.target.files[0]) {
                        setSelectedFile(e.target.files[0])
                    }
                }}
            />

            {/* Attachment Button */}
            <button
                type="button"
                onClick={() => fileInputRef.current?.click()}
                className="bg-gray-200 text-gray-600 px-4 py-3 rounded-lg hover:bg-gray-300 transition-colors flex items-center justify-center"
            >
                <Paperclip size={20} />
            </button>

            <input
                type="text"
                placeholder="Type a message..."
                className="flex-1 border border-gray-300 rounded-lg p-3 shadow-sm focus:outline-none focus:ring-2 focus:ring-purple-500"
                value={newMessage}
                onChange={(e) => setNewMessage(e.target.value)}
                disabled={isUploading}
            />

            <button 
                type="submit" 
                disabled={isUploading}
                className={`
                    bg-emerald-400 text-white px-5 py-3 rounded-lg shadow-md 
                    hover:bg-emerald-500 transition-all duration-300 flex items-center justify-center
                    ${isUploading ? 'opacity-50 cursor-not-allowed' : ''}
                `}
            >
                {isUploading ? (
                    <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
                ) : (
                    <Send size={20} />
                )}
            </button>
        </form>
      </div>
    </div>
  )
}