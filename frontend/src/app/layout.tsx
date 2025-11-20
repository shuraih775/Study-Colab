import './globals.css'
import { ReactNode } from 'react'
import Navbar from '@/components/Navbar'
import { cookies } from "next/headers";
import { getUserFromSession } from "@/libs/auth";
import { UserProvider } from '@/context/UserProvider';


export const metadata = {
  title: 'Study Buddy',
  description: 'Your study group organizer',
}

export default async  function RootLayout({ children }: { children: ReactNode }) {

  const token = cookies().get("token")?.value;
  const user = token ? await getUserFromSession(token) : null;
  return (
    <html lang="en">
      <body>
        {/* <Navbar/> */}
        
        <UserProvider user={user}>
          {children}
        </UserProvider>
      </body>
    </html>
  )
}
