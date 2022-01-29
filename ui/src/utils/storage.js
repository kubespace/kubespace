const storage={
  set(key,value){
     localStorage.setItem(key,JSON.stringify(value))   
  },
  get(key){
    if (localStorage.getItem(key)) 
      return JSON.parse(localStorage.getItem(key))
    return null
  },
  remove(key){
    localStorage.removeItem(key)
  }
}

export default storage