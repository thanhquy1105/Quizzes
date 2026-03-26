import { useState, useEffect, useRef, useCallback } from 'react';
import { WkClient, MsgType } from './WkClient';
import { Trophy, Users, Send, Zap, LogIn, Crown, Activity, AlertCircle, RefreshCw } from 'lucide-react';
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

export default function App() {
  const [name, setName] = useState('');
  const [quizId, setQuizId] = useState('BTASKEE-QUIZ-2024');
  const [uid] = useState('user-' + Math.floor(Math.random() * 1000));
  const [joined, setJoined] = useState(false);
  const [leaderboard, setLeaderboard] = useState<Participant[]>([]);
  const [lastScoreChange, setLastScoreChange] = useState<string | null>(null);
  const [status, setStatus] = useState<'disconnected' | 'connecting' | 'connected' | 'error'>('disconnected');
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  
  const clientRef = useRef<WkClient | null>(null);

  const connect = useCallback(async () => {
    if (!name) return;
    
    setStatus('connecting');
    setErrorMsg(null);
    
    const client = new WkClient('ws://localhost:8082/ws');
    // const client = new WkClient('ws://localhost:8081');
    try {
      await client.connect();
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

      // Send join request
      await client.request('/join', {
        quiz_id: quizId,
        uid: uid,
        name: name
      });

      setJoined(true);
      console.log("join true")
    } catch (err) {
      console.error(err);
      setStatus('error');
      setErrorMsg("Failed to connect to Quiz Server. Please check if the server is running on port 8082.");
    }
  }, [name, quizId, uid]);

  useEffect(() => {
    return () => {
      clientRef.current?.close();
    };
  }, []);

  const submitScore = async () => {
    if (!clientRef.current) return;
    try {
        await clientRef.current.request('/answer', {
          quiz_id: quizId,
          uid: uid,
          is_correct: true
        });
        setLastScoreChange(uid);
        setTimeout(() => setLastScoreChange(null), 800);
    } catch (err) {
        console.error("Failed to submit score", err);
    }
  };

  return (
    <div className="min-h-screen bg-[#020617] text-slate-100 p-4 font-sans selection:bg-blue-500/30 selection:text-blue-200 overflow-x-hidden relative">
      {/* Background Decorative Elements */}
      <div className="fixed top-[-10%] left-[-10%] w-[40%] h-[40%] bg-blue-600/10 blur-[120px] rounded-full pointer-events-none" />
      <div className="fixed bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-emerald-600/10 blur-[120px] rounded-full pointer-events-none" />

      <div className="max-w-6xl mx-auto w-full min-h-[90vh] flex flex-col items-center justify-center relative z-10 py-8">
        {!joined ? (
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
                  <label className="text-sm font-bold text-slate-500 uppercase tracking-[0.2em] ml-1">Session Protocol</label>
                  <div className="relative group">
                    <div className="absolute inset-0 bg-blue-500/5 rounded-2xl blur-lg transition-opacity opacity-0 group-focus-within:opacity-100" />
                    <input 
                      value={quizId}
                      onChange={e => setQuizId(e.target.value)}
                      className="relative w-full bg-slate-900/50 border border-white/5 rounded-2xl px-5 py-4 outline-none focus:border-blue-500/50 transition-all font-mono text-blue-300"
                      placeholder="Enter Session ID..."
                    />
                  </div>
                </div>

                <div className="space-y-2">
                  <label className="text-sm font-bold text-slate-500 uppercase tracking-[0.2em] ml-1">Your Alias</label>
                  <div className="relative group">
                    <div className="absolute inset-0 bg-blue-500/5 rounded-2xl blur-lg transition-opacity opacity-0 group-focus-within:opacity-100" />
                    <input 
                      value={name}
                      onChange={e => setName(e.target.value)}
                      autoFocus
                      onKeyDown={(e) => e.key === 'Enter' && name && connect()}
                      className="relative w-full bg-slate-900/50 border border-white/5 rounded-2xl px-5 py-4 outline-none focus:border-blue-500/50 transition-all text-lg"
                      placeholder="What should we call you?"
                    />
                  </div>
                </div>
              </div>

              <button 
                onClick={connect}
                disabled={!name || status === 'connecting'}
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
        ) : (
          <div className="flex flex-col lg:row gap-10 w-full animate-in fade-in zoom-in-95 duration-700">
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 items-start">
              {/* Main Game Interface */}
              <div className="lg:col-span-2 space-y-8">
                <div className="glass p-10 rounded-[2.5rem] relative overflow-hidden flex flex-col gap-10">
                   <div className="absolute -top-12 -right-12 p-4 opacity-[0.03] text-blue-400 rotate-12">
                     <Zap size={240} />
                   </div>
                   
                   <div className="flex justify-between items-start">
                      <div className="space-y-1">
                         <div className="flex items-center gap-2 text-blue-400 uppercase tracking-[0.2em] font-black text-xs">
                            <Activity size={12} className="animate-pulse" />
                            Live Session
                         </div>
                         <h2 className="text-4xl font-black tracking-tight">{quizId}</h2>
                      </div>
                      
                      <div className="px-5 py-2.5 rounded-2xl bg-white/5 border border-white/10 backdrop-blur-md flex items-center gap-3">
                         <div className="w-2.5 h-2.5 rounded-full bg-emerald-500 shadow-[0_0_12px_rgba(16,185,129,0.5)] animate-pulse" />
                         <span className="text-sm font-bold text-slate-300 tracking-wide uppercase">Connected</span>
                      </div>
                   </div>

                   <div className="bg-slate-900/60 border border-white/5 rounded-[2rem] p-10 flex flex-col gap-8 relative group">
                      <div className="space-y-4">
                        <h4 className="text-2xl font-bold text-white leading-tight">Ready for the challenge?</h4>
                        <p className="text-slate-400 text-lg leading-relaxed max-w-md">Simulate a correct answer to climb the leaderboard in real-time. Every correct strike earns you points.</p>
                      </div>
                      <button 
                        onClick={submitScore}
                        className="w-full h-24 bg-gradient-to-r from-emerald-500 via-teal-500 to-emerald-500 bg-[length:200%_auto] hover:bg-right transition-all duration-500 text-slate-950 font-[950] text-2xl rounded-2xl shadow-[0_20px_50px_rgba(16,185,129,0.25)] active:scale-[0.98] flex items-center justify-center gap-4 relative overflow-hidden"
                      >
                        <div className="absolute inset-0 bg-white/10 opacity-0 group-hover:opacity-100 transition-opacity" />
                        SUBMIT STRIKE <Send size={28} strokeWidth={3} />
                      </button>
                   </div>

                   <div className="flex flex-wrap items-center gap-8 text-slate-400 pt-4 border-t border-white/5 mt-auto">
                      <div className="flex items-center gap-2.5 font-bold text-slate-200">
                        <div className="p-2 rounded-xl bg-blue-500/10 text-blue-400"><Users size={20}/></div>
                        <span className="text-lg">{leaderboard.length} <span className="text-slate-500 font-medium ml-1">Elite Players</span></span>
                      </div>
                      <div className="h-8 w-px bg-white/5" />
                      <div className="text-slate-500 font-medium">Session ID: <span className="font-mono text-blue-400/80 ml-1">{uid}</span></div>
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
