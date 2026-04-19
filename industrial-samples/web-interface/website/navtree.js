function highlight(current){ if(highlighted!=0){overC(highlighted,"#bcbcbc");} overC(current,"#dedede"); highlighted=current; if(eval("typeof(parent.timeoutPntr)")!='undefined')parent.window.clearTimeout(parent.timeoutPntr); }
function showBranch(branch){ var objBranch=document.getElementById(branch).style; if (objBranch.display=="block"){objBranch.display="none";}else {objBranch.display="block";} top.frames.home.focus();}
function branchState(branch,state){ objBranch=document.getElementById(branch).style; if(state=="expand"){objBranch.display="block";}else{objBranch.display="none";} }
function swapFolder(img){ objImg=document.getElementById(img); if(objImg.src.indexOf('folder.gif') > -1){objImg.src=openImg.src;} else{objImg.src=closedImg.src;} }
function folderState(img,state){ objImg=document.getElementById(img); if(state=="expand"){objImg.src=openImg.src;} else{objImg.src=closedImg.src;}}
function overC(object, color){ object.style.backgroundColor=color; }
function outC(object,color){ if(object==highlighted){color="#dedede";} object.style.backgroundColor=color; }
function changeAll(state){ folderState("f_data",state); folderState("f_diag",state); folderState("f_admin",state); branchState("b_data",state); branchState("b_diag",state); branchState("b_admin",state); top.frames.home.focus(); }