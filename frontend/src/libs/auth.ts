import axios from "axios"


export const getUserFromSession = async (token)=>{
    try{
        const res = await axios.get('http://localhost:8080/user/me',{
        headers: {
        Cookie: `token=${token}`
      }
        
    })
    console.log(res.data)
        return res.data;
    }
    catch(err){
console.log(err);
    }
    
    
    
     
}