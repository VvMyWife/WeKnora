async(page)=>{
 const session=await page.context().newCDPSession(page)
 await session.send('Network.clearBrowserCache')
 await session.detach()
 await page.reload()
}
