export interface User {
  id: string
  stravaId: string
  name: string
  email: string
  profilePicture?: string
  createdAt: string
  updatedAt: string
}

export interface Session {
  id: string
  title: string
  createdAt: string
  updatedAt: string
}

export interface Message {
  id: string
  sessionId: string
  role: 'user' | 'assistant'
  content: string
  timestamp: string
}

export interface Activity {
  id: string
  stravaId: string
  name: string
  type: string
  distance: number
  movingTime: number
  totalElevationGain: number
  startDate: string
  averageSpeed?: number
  maxSpeed?: number
  averageHeartrate?: number
  maxHeartrate?: number
  calories?: number
}

export interface LogbookEntry {
  id: string
  userId: string
  date: string
  notes?: string
  mood?: number
  energy?: number
  sleep?: number
  stress?: number
  createdAt: string
  updatedAt: string
}