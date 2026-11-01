// Canvas drawing code is ported from the approved TokenInfra production artwork.
// @ts-nocheck
/* oxlint-disable no-unused-expressions */
import { useEffect, useRef } from 'react'

import { useTranslation } from 'react-i18next'

export function TokenInfraMark({ className }: { className?: string }) {
  return (
    <svg viewBox='0 0 40 36' fill='none' aria-hidden='true' className={className}>
      <path d='M3 29 17 6a4 4 0 0 1 7 0l14 23H27L20.5 18 14 29Z' fill='currentColor' />
      <path d='m17 29 3.5-6 3.5 6Z' fill='currentColor' />
    </svg>
  )
}

export function TokenFactoryVisual() {
  const { t } = useTranslation()
  const canvasRef = useRef<HTMLCanvasElement>(null)

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    if (navigator.userAgent.toLowerCase().includes('jsdom')) return
    const ctx = canvas.getContext('2d')
    if (!ctx) return
    {let w=660,h=610,t=0,last=0,frame:number|undefined,visible=true;const reduced=matchMedia('(prefers-reduced-motion: reduce)').matches;function resize(){const d=Math.min(devicePixelRatio||1,2);w=canvas.clientWidth;h=canvas.clientHeight;canvas.width=w*d;canvas.height=h*d;ctx.setTransform(d,0,0,d,0,0)}new ResizeObserver(resize).observe(canvas);resize();const P=(x,y,z)=>[330+(x-y)*1.14,300+(x+y)*.47-z];function line(points,color,width=1){ctx.beginPath();points.forEach((p,i)=>{const q=P(...p);i?ctx.lineTo(...q):ctx.moveTo(...q)});ctx.strokeStyle=color;ctx.lineWidth=width;ctx.stroke()}function face(points,fill,stroke){ctx.beginPath();points.forEach((p,i)=>{const q=P(...p);i?ctx.lineTo(...q):ctx.moveTo(...q)});ctx.closePath();ctx.fillStyle=fill;ctx.fill();if(stroke){ctx.strokeStyle=stroke;ctx.lineWidth=.8;ctx.stroke()}}function box(x,y,z,sx,sy,sz,c){face([[x,y,z],[x+sx,y,z],[x+sx,y,z+sz],[x,y,z+sz]],c[0],c[3]);face([[x+sx,y,z],[x+sx,y+sy,z],[x+sx,y+sy,z+sz],[x+sx,y,z+sz]],c[1],c[3]);face([[x,y,z+sz],[x+sx,y,z+sz],[x+sx,y+sy,z+sz],[x,y+sy,z+sz]],c[2],c[3])}const steel=['#101d35','#10152b','#1d2945','#6689b570'];const token=['#27445e','#233752','#699ac1','#a5d7ecb0'];function draw(now){frame=undefined;if(!visible)return;if(!reduced)t+=Math.min(now-last||16,40)*.001;last=now;ctx.clearRect(0,0,w,h);ctx.save();const scale=Math.min(w/660,h/610);ctx.translate((w-660*scale)/2,(h-610*scale)/2);ctx.scale(scale,scale);const g=ctx.createRadialGradient(330,290,0,330,290,330);g.addColorStop(0,'#354c9f40');g.addColorStop(1,'#18225100');ctx.fillStyle=g;ctx.fillRect(0,0,660,610);
for(let i=-220;i<=220;i+=28){line([[i,-220,-34],[i,220,-34]],'#566ea315',.5);line([[-220,i,-34],[220,i,-34]],'#566ea315',.5)}
// Holographic floor, rotating perimeter and processing satellites.
for(let r=0;r<3;r++){const radius=110+r*24,points=[];for(let k=0;k<=96;k++){const a=k*Math.PI/48;points.push([Math.cos(a)*radius,Math.sin(a)*radius,-32])}line(points,['#7297ff50','#7697df25','#8c7fe91f'][r],.7);const a=t*(r%2?-.35:.3)+r*2;ctx.save();ctx.shadowColor='#a2cfff';ctx.shadowBlur=12;line([[Math.cos(a)*radius,Math.sin(a)*radius,-31],[Math.cos(a+.22)*radius,Math.sin(a+.22)*radius,-31],[Math.cos(a+.44)*radius,Math.sin(a+.44)*radius,-31]],'#9dceff',1.7);ctx.restore()}
for(let k=0;k<42;k++){const a=k*2.399,rad=170+(k%5)*12,z=(t*11+k*9)%170;const p=P(Math.cos(a)*rad,Math.sin(a)*rad,z-25);ctx.fillStyle=`rgba(139,161,242,${.08+Math.sin(z/170*Math.PI)*.3})`;ctx.fillRect(p[0],p[1],k%6===0?2:1,k%6===0?2:1)}
for(const [x,y] of [[-110,98],[90,-115]]){line([[x,y,-26],[x,y,28],[x*.65,y*.65,28]],'#729ce955',.8);box(x-13,y-13,-28,26,26,12,steel);box(x-8,y-8,-12,16,16,10,['#2e3b68','#202e50','#7286c5','#99b3edaa']);const p=P(x,y,10);ctx.fillStyle='#92bbdd';ctx.fillRect(p[0]-2,p[1]-2,4,4)}
// Elevated transport ribbons: incoming requests and outgoing tokens.
for(const y of [-58,44]){face([[-240,y,0],[-72,y,0],[-72,y+22,0],[-240,y+22,0]],'#101d32','#526d9b66');line([[-240,y+11,1],[-70,y+11,1]],'#91a8ed40');const sweep=-235+(t*64%160);ctx.save();ctx.shadowColor='#7bbfe4';ctx.shadowBlur=8;line([[sweep-20,y+11,2],[sweep,y+11,2]],'#a5d7ed',1.5);ctx.restore();for(let k=0;k<7;k++){const x=-240+((t*32+k*28)%168);const p=P(x,y+11,4);ctx.fillStyle=k%2?'#8ba4dc':'#8ecbd4';ctx.globalAlpha=.6;ctx.fillRect(p[0]-2,p[1]-2,4,4);ctx.globalAlpha=1}}
face([[65,-10,0],[238,-10,0],[238,34,0],[65,34,0]],'#13233b','#6482b377');for(let i=0;i<12;i++)line([[74+i*14,-10,1],[74+i*14,34,1]],'#708fbe28',.7);
// Layered foundry pedestal and translucent model processing core.
box(-70,-70,-30,140,140,14,steel);box(-62,-62,-16,124,124,12,['#182640','#151e37','#273353','#7994c176']);box(-48,-48,-4,96,96,25,['#14283d','#1b2549','#25385c','#82a7d5a0']);
for(let j=0;j<4;j++){const z=38+j*25+Math.sin(t*1.2+j*.5)*3;face([[-42,-42,z],[42,-42,z],[42,42,z],[-42,42,z]],['#3b659c20','#6478c921','#749ccc18','#839fea12'][j],'#91baf095');line([[-31,-31,z+1],[20,-31,z+1]],'#b6d2fa80',1.2);line([[-31,-21,z+1],[-6,-21,z+1]],'#8db7ec50',.8)}
for(const x of [-40,40])for(const y of [-40,40])line([[x,y,20],[x,y,112]],'#8eb3e13d',.8);
const pulse=(Math.sin(t*1.8)+1)/2;const q=P(0,0,63);const glow=ctx.createRadialGradient(q[0],q[1],0,q[0],q[1],45);glow.addColorStop(0,`rgba(74,203,185,${.2+pulse*.12})`);glow.addColorStop(1,'rgba(74,203,185,0)');ctx.fillStyle=glow;ctx.fillRect(q[0]-45,q[1]-45,90,90);box(-9,-9,51,18,18,18,['#167c7a','#22485f','#78c9c2','#b2f3e9']);
// Soft vertical energy column and animated scan plane.
const top=P(0,0,178),bottom=P(0,0,20);const beam=ctx.createLinearGradient(top[0],top[1],bottom[0],bottom[1]);beam.addColorStop(0,'#8fb6ff00');beam.addColorStop(.55,'#91bffb12');beam.addColorStop(1,'#6de2d926');ctx.fillStyle=beam;ctx.beginPath();ctx.moveTo(top[0]-10,top[1]);ctx.lineTo(top[0]+10,top[1]);ctx.lineTo(bottom[0]+30,bottom[1]);ctx.lineTo(bottom[0]-30,bottom[1]);ctx.fill();
const scan=25+(t*22%110);ctx.save();ctx.shadowBlur=12;ctx.shadowColor='#7eafff';face([[-47,-47,scan],[47,-47,scan],[47,47,scan],[-47,47,scan]],'#8ec8ff0c','#a0c9ff8a');ctx.restore();
for(let j=0;j<2;j++){const pts=[],radius=75+j*12;for(let k=0;k<=90;k++){const a=k*Math.PI/45+t*(j?-.25:.25);pts.push([Math.cos(a)*radius,Math.sin(a)*radius,100+j*18+Math.sin(a)*13])}line(pts,j?'#ba9ffb59':'#92ccf58a',.8);const a=t*(j?-1:1)+j*3,p=P(Math.cos(a)*radius,Math.sin(a)*radius,100+j*18+Math.sin(a)*13);ctx.save();ctx.shadowBlur=14;ctx.shadowColor=j?'#ae93ff':'#a2dfff';ctx.fillStyle=j?'#dbcfff':'#c2eeff';ctx.beginPath();ctx.arc(p[0],p[1],2.2,0,Math.PI*2);ctx.fill();ctx.restore()}
for(let k=0;k<5;k++){const x=78+((t*26+k*33)%158),z=3+Math.sin(t*1.2+k)*1.5;ctx.save();ctx.shadowColor='#79b8ed';ctx.shadowBlur=13;line([[x-11,9,z+2],[x,9,z+2]],'#87c5f8a0',1.5);box(x,3,z,13,13,13,token);ctx.restore()}
// Sparse output particles dissolve at the end of the belt.
for(let k=0;k<14;k++){const a=(t*.35+k*.071)%1,p=P(238+a*22,10+(k%3-1)*10,10+a*70);ctx.fillStyle=`rgba(149,184,241,${(1-a)*.6})`;ctx.fillRect(p[0],p[1],1.5,1.5)}
ctx.restore();if(!reduced)frame=requestAnimationFrame(draw)}const observer = new IntersectionObserver((entries) => {
      visible = entries[0]?.isIntersecting ?? true
      if (frame) cancelAnimationFrame(frame)
      if (visible) { last = 0; draw(performance.now()) }
    })
    observer.observe(canvas)
    const onVisibilityChange = () => {
      if (document.hidden) { if (frame) cancelAnimationFrame(frame) }
      else if (visible) { last = 0; draw(performance.now()) }
    }
    document.addEventListener('visibilitychange', onVisibilityChange)
    draw(performance.now())
    return () => {
      if (frame) cancelAnimationFrame(frame)
      observer.disconnect()
      document.removeEventListener('visibilitychange', onVisibilityChange)
    }
  }

  }, [])

  return (
    <div className='ti-orb' role='img' aria-label={t('Requests flow into TokenInfra and leave as usable tokens')}>
      <canvas ref={canvasRef} className='ti-orb-canvas' aria-hidden='true' />
    </div>
  )
}
