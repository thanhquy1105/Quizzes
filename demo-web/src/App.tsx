import { useState, useEffect, useRef, useCallback } from 'react';
import { WsClient, MsgType } from './WsClient';
import { Trophy, Users, Zap, LogIn, Crown, Activity, AlertCircle, RefreshCw } from 'lucide-react';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

interface Participant {
  uid: string;
  name: string;
  score: number;
}

interface Answer {
  ID: number;
  Content: string;
  IsCorrect: boolean;
}

interface Question {
  ID: number;
  Content: string;
  Answers: Answer[];
  answered: boolean;
}

interface Quiz {
  ID: number;
  Title: string;
  Description: string;
  question_count: number;
  participant_count: number;
  Questions?: Question[];
}

interface QuizSession {
  ID: number;
  QuizID: number;
  SessionCode: string;
  Name: string;
  CreatedAt: string;
}

function SocketStatus({ status }: { status: 'disconnected' | 'connecting' | 'connected' | 'error' }) {
  const statusConfig = {
    connected: { color: 'bg-emerald-500', text: 'Connected', shadow: 'shadow-[0_0_12px_rgba(16,185,129,0.5)]' },
    connecting: { color: 'bg-blue-500', text: 'Connecting...', shadow: 'shadow-[0_0_12px_rgba(59,130,246,0.5)]' },
    disconnected: { color: 'bg-slate-500', text: 'Disconnected', shadow: '' },
    error: { color: 'bg-red-500', text: 'Conn Error', shadow: 'shadow-[0_0_12px_rgba(239,68,68,0.5)]' },
  };

  const current = statusConfig[status];

  return (
    <div className="px-4 py-2 rounded-xl bg-white/5 border border-white/10 backdrop-blur-md flex items-center gap-2.5 transition-all duration-500">
      <div className={cn(
        "w-2 h-2 rounded-full",
        current.color,
        current.shadow,
        (status === 'connecting' || status === 'connected') && "animate-pulse"
      )} />
      <span className="text-[10px] font-black text-slate-300 tracking-[0.1em] uppercase">{current.text}</span>
    </div>
  );
}

export default function App() {
  const [name, setName] = useState('');
  const [username, setUsername] = useState('');
  const [quizId, setQuizId] = useState('');
  const [uid, setUid] = useState('');
  const [joined, setJoined] = useState(false);
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [quizzes, setQuizzes] = useState<Quiz[]>([]);
  const [sessions, setSessions] = useState<QuizSession[]>([]);
  const [leaderboard, setLeaderboard] = useState<Participant[]>([]);
  const [lastScoreChange, setLastScoreChange] = useState<string | null>(null);
  const [status, setStatus] = useState<'disconnected' | 'connecting' | 'connected' | 'error'>('disconnected');
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [quizError, setQuizError] = useState<string | null>(null);
  const [activeQuiz, setActiveQuiz] = useState<Quiz | null>(null);
  const [activeSessionCode, setActiveSessionCode] = useState<string>('');
  const clientRef = useRef<WsClient | null>(null);

  const [accessToken, setAccessToken] = useState('');
  const [refreshToken, setRefreshToken] = useState('');

  const refreshSession = useCallback(async () => {
    if (!refreshToken) return null;
    try {
      const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/refresh`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });

      if (response.ok) {
        const data = await response.json();
        setAccessToken(data.access_token);
        setRefreshToken(data.refresh_token);
        console.log("Session refreshed successfully");
        return data.access_token as string;
      } else {
        console.warn("Refresh token expired or invalid");
        setIsLoggedIn(false);
        setJoined(false);
        setStatus('disconnected');
        setErrorMsg("Session expired. Please login again.");
        return null;
      }
    } catch (err) {
      console.error("Refresh failed", err);
      return null;
    }
  }, [refreshToken]);

  const callApi = useCallback(async (url: string, options: RequestInit = {}, tokenOverride?: string) => {
    const headers = new Headers(options.headers);
    const token = tokenOverride || accessToken;
    if (token) {
      headers.set('Authorization', `Bearer ${token}`);
    }

    let response = await fetch(url, { ...options, headers });

    if (response.status === 401) {
      console.log("Access token expired, attempting refresh...");
      const newToken = await refreshSession();
      if (newToken) {
        headers.set('Authorization', `Bearer ${newToken}`);
        response = await fetch(url, { ...options, headers });
      }
    }

    return response;
  }, [accessToken, refreshSession]);

  const connect = useCallback(async () => {
    if (!name || !username) return;

    setStatus('connecting');
    setErrorMsg(null);

    try {
      // 1. Login request (HTTP)
      const loginResponse = await fetch(`${import.meta.env.VITE_API_BASE_URL}/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          username: username,
          name: name
        }),
      });

      if (!loginResponse.ok) {
        throw new Error(`Login failed with status: ${loginResponse.status}`);
      }

      const loginData = await loginResponse.json();
      const initialToken = loginData.access_token;
      setAccessToken(initialToken);
      setRefreshToken(loginData.refresh_token);
      setUid(username);

      // 2. Fetch Quizzes (HTTP with Auth & Auto-Refresh)
      const quizResponse = await callApi(`${import.meta.env.VITE_API_BASE_URL}/quizzes`, {}, initialToken);

      if (!quizResponse.ok) {
        throw new Error(`Failed to fetch quizzes: ${quizResponse.status}`);
      }

      const quizData = await quizResponse.json();
      setQuizzes(quizData.quizzes || []);

      // 3. Fetch Sessions (HTTP)
      const sessionResponse = await callApi(`${import.meta.env.VITE_API_BASE_URL}/sessions`, {}, initialToken);
      if (sessionResponse.ok) {
        const sessionData = await sessionResponse.json();
        setSessions(sessionData.sessions || []);
      }

      setIsLoggedIn(true);

      // 3. Connect to WebSocket (with Token Authentication & Retry on Refresh)
      const client = new WsClient(import.meta.env.VITE_WS_URL);

      try {
        await client.connect(username, initialToken);
      } catch (wsErr) {
        console.log("WebSocket initial auth failed, trying refresh...");
        const refreshedToken = await refreshSession();
        if (refreshedToken) {
          await client.connect(username, refreshedToken);
        } else {
          throw wsErr;
        }
      }

      clientRef.current = client;
      setStatus('connected');

      // Handle message updates
      client.onMessage((msgType, data) => {
        if (msgType === MsgType.Message) {
          try {
            const text = new TextDecoder().decode(data);
            const msgObj = JSON.parse(text);
            if (msgObj.type === 'leaderboard') {
              setLeaderboard(msgObj.leaderboard);
            }
          } catch (e) {
            console.error("Failed to parse message", e);
          }
        }
      });
    } catch (err) {
      console.error(err);
      setStatus('error');
      setErrorMsg(err instanceof Error ? err.message : "Failed to connect or login to Quiz Server.");
    }
  }, [name, username]);

  const joinQuiz = async (sessionCode: string) => {
    if (!clientRef.current) return;
    try {
      setActiveSessionCode(sessionCode);
      const quizResult = await clientRef.current.request('/join', {
        session_code: sessionCode,
      });

      const fullQuiz = quizResult.body;
      setActiveQuiz(fullQuiz);
      setJoined(true);
    } catch (err: any) {
      console.error("Failed to join quiz", err);
      const msg = err.message || "Failed to join the selected quiz.";
      setErrorMsg(msg);
    }
  };

  useEffect(() => {
    return () => {
      clientRef.current?.close();
    };
  }, []);

  const submitAnswer = async (questionId: number, answerId: number) => {
    if (!clientRef.current) return;
    try {
      await clientRef.current.request('/answer', {
        session_code: activeSessionCode,
        question_id: questionId,
        answer_id: answerId,
      });
      setLastScoreChange(uid);
      setTimeout(() => setLastScoreChange(null), 800);
      // Mark question as answered locally so it disappears from the list
      setActiveQuiz(prev => {
        if (!prev?.Questions) return prev;
        return {
          ...prev,
          Questions: prev.Questions.map(q =>
            q.ID === questionId ? { ...q, answered: true } : q
          ),
        };
      });
    } catch (err: any) {
      console.error("Failed to submit answer", err);
      const msg = err?.message || 'Failed to submit answer';
      setQuizError(msg);
      setTimeout(() => setQuizError(null), 4000);
    }
  };

  const currentQuiz = quizzes.find(q => q.ID.toString() === quizId);

  return (
    <div className="min-h-screen bg-[#020617] text-slate-100 p-4 font-sans selection:bg-blue-500/30 selection:text-blue-200 overflow-x-hidden relative">
      {/* Background Decorative Elements */}
      <div className="fixed top-[-10%] left-[-10%] w-[40%] h-[40%] bg-blue-600/10 blur-[120px] rounded-full pointer-events-none" />
      <div className="fixed bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-emerald-600/10 blur-[120px] rounded-full pointer-events-none" />

      <div className="max-w-6xl mx-auto w-full min-h-[90vh] flex flex-col items-center justify-center relative z-10 py-8">
        {!isLoggedIn ? (
          <div className="w-full max-w-lg animate-in fade-in slide-in-from-bottom-8 duration-1000">
            <div className="glass p-10 rounded-[2.5rem] border-white/10 flex flex-col gap-8">
              <div className="flex flex-col items-center text-center gap-4">
                <div className="relative">
                  <div className="absolute inset-0 bg-blue-500 blur-2xl opacity-20 animate-pulse" />
                  <div className="relative p-5 rounded-3xl bg-gradient-to-br from-blue-500/20 to-blue-600/5 text-blue-400 border border-blue-500/20 shadow-2xl">
                    <Zap size={48} strokeWidth={2.5} />
                  </div>
                </div>
                <div>
                  <h1 className="text-4xl font-black tracking-tight mb-2">
                    <span className="bg-clip-text text-transparent bg-gradient-to-r from-blue-400 via-indigo-400 to-blue-400 bg-[length:200%_auto] animate-gradient">QUIZ MASTER</span>
                  </h1>
                  <p className="text-slate-400 text-lg font-medium">Elevate your knowledge in real-time</p>
                </div>
              </div>

              {status === 'error' && (
                <div className="bg-red-500/10 border border-red-500/20 rounded-2xl p-4 flex items-start gap-3 animate-in fade-in zoom-in-95">
                  <AlertCircle className="text-red-400 shrink-0 mt-0.5" size={20} />
                  <div className="flex flex-col gap-1">
                    <p className="text-sm font-semibold text-red-200">Connection Error</p>
                    <p className="text-xs text-red-400/80 leading-relaxed">{errorMsg}</p>
                  </div>
                </div>
              )}

              <div className="space-y-6">
                <div className="space-y-2">
                  <label className="text-sm font-bold text-slate-500 uppercase tracking-[0.2em] ml-1">Username</label>
                  <div className="relative group">
                    <div className="absolute inset-0 bg-blue-500/5 rounded-2xl blur-lg transition-opacity opacity-0 group-focus-within:opacity-100" />
                    <input
                      value={username}
                      onChange={e => setUsername(e.target.value)}
                      className="relative w-full bg-slate-900/50 border border-white/5 rounded-2xl px-5 py-4 outline-none focus:border-blue-500/50 transition-all font-mono text-blue-300"
                      placeholder="Enter your username..."
                    />
                  </div>
                </div>

                <div className="space-y-2">
                  <label className="text-sm font-bold text-slate-500 uppercase tracking-[0.2em] ml-1">Display Name</label>
                  <div className="relative group">
                    <div className="absolute inset-0 bg-blue-500/5 rounded-2xl blur-lg transition-opacity opacity-0 group-focus-within:opacity-100" />
                    <input
                      value={name}
                      onChange={e => setName(e.target.value)}
                      onKeyDown={(e) => e.key === 'Enter' && name && username && connect()}
                      className="relative w-full bg-slate-900/50 border border-white/5 rounded-2xl px-5 py-4 outline-none focus:border-blue-500/50 transition-all text-lg"
                      placeholder="How should we call you?"
                    />
                  </div>
                </div>
              </div>

              <button
                onClick={connect}
                disabled={!name || !username || status === 'connecting'}
                className={cn(
                  "relative group overflow-hidden w-full h-16 rounded-2xl font-bold text-lg transition-all",
                  status === 'connecting' ? "bg-slate-800 pointer-events-none" : "bg-white text-slate-950 active:scale-95 hover:shadow-[0_0_40px_rgba(255,255,255,0.15)]"
                )}
              >
                <div className="relative z-10 flex items-center justify-center gap-3">
                  {status === 'connecting' ? (
                    <RefreshCw className="animate-spin" size={24} />
                  ) : (
                    <>
                      ENTER ARENA <LogIn size={20} className="group-hover:translate-x-1.5 transition-transform" />
                    </>
                  )}
                </div>
              </button>
            </div>
          </div>
        ) : !joined ? (
          <div className="w-full max-w-4xl space-y-12 animate-in fade-in slide-in-from-bottom-8 duration-700">
            {/* Active Sessions Section */}
            {sessions.length > 0 && (
              <div className="glass p-10 rounded-[2.5rem] border-white/5 flex flex-col gap-8 bg-blue-500/5">
                <div className="flex items-center justify-between">
                  <div className="space-y-1">
                    <div className="flex items-center gap-2 text-blue-400 uppercase tracking-[0.2em] font-black text-[10px]">
                      <Activity size={12} className="animate-pulse" />
                      Live Sessions
                    </div>
                    <h2 className="text-3xl font-black tracking-tight">QUIZ SESSIONS</h2>
                  </div>
                  <div className="flex items-center gap-4">
                    {errorMsg && (
                       <div className="px-4 py-2 rounded-xl bg-red-500/10 border border-red-500/20 flex items-center gap-2 text-red-400 animate-in fade-in slide-in-from-right-4">
                          <AlertCircle size={14} />
                          <span className="text-[10px] font-bold uppercase tracking-wider">{errorMsg}</span>
                       </div>
                    )}
                    <SocketStatus status={status} />
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {sessions.map(session => {
                    const quiz = quizzes.find(q => q.ID === session.QuizID);
                    return (
                      <div key={session.ID} className="p-5 rounded-2xl bg-white/5 border border-white/10 hover:border-blue-500/30 transition-all group/session">
                        <div className="flex flex-col gap-3">
                          <div className="flex justify-between items-start">
                            <span className="text-xs font-mono text-blue-400/70">{session.SessionCode}</span>
                            <span className="px-2 py-0.5 rounded-md bg-emerald-500/10 text-emerald-400 text-[10px] font-black uppercase">Active</span>
                          </div>
                          <h4 className="font-bold text-white group-hover/session:text-blue-300 transition-colors uppercase tracking-tight">{session.Name || `Session #${session.ID}`}</h4>
                          <p className="text-slate-500 text-xs font-medium">{quiz?.Title || 'General Quiz'}</p>
                          <button
                            onClick={() => joinQuiz(session.SessionCode)}
                            className="mt-2 w-full py-2.5 rounded-xl bg-blue-500/10 hover:bg-blue-500 text-blue-400 hover:text-white text-xs font-black transition-all uppercase tracking-widest border border-blue-500/20"
                          >
                            Join Session
                          </button>
                        </div>
                      </div>
                    )
                  })}
                </div>
              </div>
            )}


          </div>
        ) : (
          <div className="flex flex-col lg:row gap-10 w-full animate-in fade-in zoom-in-95 duration-700">
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 items-start">
              {/* Main Game Interface */}
              <div className="lg:col-span-2 space-y-8">
                <div className="glass p-10 rounded-[2.5rem] relative overflow-hidden flex flex-col gap-10">
                  <div className="absolute -top-12 -right-12 p-4 opacity-[0.03] text-blue-400 rotate-12">
                    <Zap size={240} />
                  </div>

                  <div className="flex justify-between items-center bg-slate-900/40 p-4 rounded-2xl border border-white/5">
                    <div className="space-y-1">
                      <div className="flex items-center gap-2 text-blue-400 uppercase tracking-[0.2em] font-black text-[10px]">
                        <Activity size={12} className="animate-pulse" />
                        Live Session
                      </div>
                      <h2 className="text-3xl font-black tracking-tight">{activeQuiz?.Title || quizId}</h2>
                      <h3 className='text-2xs'>{activeQuiz?.Description}</h3>
                    </div>

                    <SocketStatus status={status} />
                  </div>

                  {/* In-quiz error banner */}
                  {quizError && (
                    <div className="flex items-center gap-3 bg-red-500/10 border border-red-500/20 rounded-2xl px-5 py-3 animate-in fade-in slide-in-from-top-2 duration-300">
                      <AlertCircle className="text-red-400 shrink-0" size={18} />
                      <span className="text-sm font-semibold text-red-300 flex-1">{quizError}</span>
                      <button onClick={() => setQuizError(null)} className="text-red-400/60 hover:text-red-300 transition-colors text-lg leading-none">&times;</button>
                    </div>
                  )}

                  <div className="bg-slate-900/60 border border-white/5 rounded-[2rem] p-10 flex flex-col gap-8 relative group min-h-[400px]">
                    {(() => {
                      const unanswered = activeQuiz?.Questions?.filter(q => !q.answered) ?? [];
                      const total = activeQuiz?.Questions?.length ?? 0;
                      const current = unanswered[0];
                      return current ? (
                        <>
                          <div className="space-y-6">
                            <div className="flex items-center gap-3">
                              <div className="px-3 py-1 rounded-full bg-blue-500/10 border border-blue-500/20 text-blue-400 text-xs font-black uppercase tracking-widest">
                                Question {total - unanswered.length + 1} of {total}
                              </div>
                            </div>
                            <h4 className="text-3xl font-bold text-white leading-tight min-h-[4rem]">
                              {current.Content}
                            </h4>
                          </div>

                          <div className="grid grid-cols-1 gap-4">
                            {current.Answers.map((answer: any, idx: number) => (
                              <button
                                key={answer.ID}
                                onClick={() => submitAnswer(current.ID, answer.ID)}
                                className="group/btn relative w-full p-6 bg-slate-800/40 border border-white/5 rounded-2xl text-left transition-all hover:bg-blue-500/10 hover:border-blue-500/30 active:scale-[0.99] flex items-center gap-4"
                              >
                                <div className="w-10 h-10 rounded-xl bg-slate-700/50 flex items-center justify-center font-black text-slate-400 group-hover/btn:bg-blue-500/20 group-hover/btn:text-blue-400 transition-colors">
                                  {String.fromCharCode(65 + idx)}
                                </div>
                                <span className="text-lg font-medium text-slate-200 group-hover/btn:text-white transition-colors">{answer.Content}</span>
                              </button>
                            ))}
                          </div>
                        </>
                      ) : (
                        <div className="flex flex-col items-center justify-center py-12 text-center gap-6 animate-in fade-in zoom-in-95 duration-500">
                          <div className="w-24 h-24 rounded-[2rem] bg-emerald-500/10 flex items-center justify-center text-emerald-500">
                            <Trophy size={48} />
                          </div>
                          <div className="space-y-2">
                            <h4 className="text-3xl font-black text-white">Quiz Completed!</h4>
                            <p className="text-slate-400 text-lg">You've answered all questions. Check your final rank in the Hall of Fame!</p>
                          </div>
                          <button
                            onClick={() => setJoined(false)}
                            className="px-8 py-3 bg-white text-slate-950 font-bold rounded-xl hover:shadow-[0_0_30px_rgba(255,255,255,0.1)] transition-all active:scale-95"
                          >
                            BACK TO ARENA
                          </button>
                        </div>
                      );
                    })()}
                  </div>

                  <div className="flex flex-wrap items-center gap-8 text-slate-400 pt-4 border-t border-white/5 mt-auto">
                    <div className="flex items-center gap-2.5 font-bold text-slate-200">
                      <div className="p-2 rounded-xl bg-blue-500/10 text-blue-400"><Users size={20} /></div>
                      <span className="text-lg">{leaderboard.length} <span className="text-slate-500 font-medium ml-1">Elite Players</span></span>
                    </div>
                    <div className="h-8 w-px bg-white/5" />
                    <div className="text-slate-500 font-medium">USER ID: <span className="font-mono text-blue-400/80 ml-1">{uid}</span></div>
                  </div>
                </div>
              </div>

              {/* Dynamic Leaderboard */}
              <div className="space-y-6">
                <div className="flex items-center justify-between px-2">
                  <div className="flex items-center gap-3 font-black text-xl tracking-tight">
                    <Trophy className="text-amber-500 drop-shadow-[0_0_10px_rgba(245,158,11,0.3)]" size={24} />
                    HALL OF FAME
                  </div>
                </div>

                <div className="glass rounded-[2rem] border-white/10 overflow-hidden shadow-2xl relative">
                  <div className="absolute inset-x-0 top-0 h-24 bg-gradient-to-b from-blue-500/5 to-transparent pointer-events-none" />

                  {leaderboard.length === 0 ? (
                    <div className="py-24 flex flex-col items-center gap-4 text-center px-8 relative z-10">
                      <div className="w-16 h-16 rounded-3xl bg-slate-800/50 flex items-center justify-center text-slate-600">
                        <RefreshCw className="animate-spin-slow" size={32} />
                      </div>
                      <p className="text-slate-500 font-medium italic">Waiting for champions to join...</p>
                    </div>
                  ) : (
                    <div className="flex flex-col relative z-10">
                      {leaderboard.map((p, index) => (
                        <div
                          key={p.uid}
                          className={cn(
                            "flex items-center gap-5 p-5 border-b border-white/5 transition-all duration-300 relative group",
                            p.uid === uid ? "bg-blue-500/10" : "hover:bg-white/[0.02]",
                            lastScoreChange === p.uid && "after:absolute after:inset-0 after:bg-emerald-500/20 after:animate-in after:fade-in after:duration-700"
                          )}
                        >
                          <div className="w-10 h-10 shrink-0 flex items-center justify-center relative">
                            {index === 0 ? (
                              <div className="absolute inset-0 bg-amber-500/20 blur-xl rounded-full" />
                            ) : null}
                            <div className={cn(
                              "relative w-full h-full rounded-2xl flex items-center justify-center font-[950] text-sm",
                              index === 0 ? "bg-amber-400 text-slate-900" :
                                index === 1 ? "bg-slate-300 text-slate-900" :
                                  index === 2 ? "bg-amber-700/50 text-amber-200 border border-amber-600/30" :
                                    "bg-slate-800 text-slate-500"
                            )}>
                              {index === 0 ? <Crown size={18} strokeWidth={3} /> : index + 1}
                            </div>
                          </div>

                          <div className="flex-1 min-w-0">
                            <div className={cn(
                              "font-bold text-lg truncate flex items-center gap-2",
                              p.uid === uid ? "text-blue-400" : "text-white"
                            )}>
                              {p.name}
                              {p.uid === uid && (
                                <span className="bg-blue-500/20 text-[10px] text-blue-400 px-2 py-0.5 rounded-full border border-blue-500/20 tracking-[0.1em] font-black uppercase">YOU</span>
                              )}
                            </div>
                            <div className="font-mono text-[11px] text-slate-500 truncate">@{p.uid}</div>
                          </div>

                          <div className="text-right">
                            <div className="font-mono font-black text-2xl text-slate-200 tabular-nums">
                              {p.score}
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
