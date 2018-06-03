export default `mutation {
    u1: CreateUser(id: "u1", name: "Will") {
      id
      name
    }
    u2: CreateUser(id: "u2", name: "Bob") {
      id
      name
    }
    u3: CreateUser(id: "u3", name: "Jenny") {
      id
      name
    }
    u4: CreateUser(id: "u4", name: "Angie") {
      id
      name
    }
    b1: CreateBusiness(
      id: "b1"
      name: "KettleHouse Brewing Co."
      address: "313 N 1st St W"
      city: "Missoula"
      state: "MT"
    ) {
      id
      name
    }
    b2: CreateBusiness(
      id: "b2"
      name: "Imagine Nation Brewing"
      address: "1151 W Broadway St"
      city: "Missoula"
      state: "MT"
    ) {
      id
      name
    }
    b3: CreateBusiness(
      id: "b3"
      name: "Ninja Mike's"
      address: "Food Truck - Farmers Market"
      city: "Missoula"
      state: "MT"
    ) {
      id
      name
    }
    b4: CreateBusiness(
      id:"b4",
      name: "Market on Front",
      address:"201 E Front St",
      city:"Missoula",
      state:"MT",
    ) {
      id
      name
    }
    b5: CreateBusiness(
      id:"b5",
      name:"Missoula Public Library",
      address:"301 E Main St",
      city: "Missoula",
      state: "MT"
    ) {
      id
      name
    }
    b6:CreateBusiness(
      id:"b6",
      name: "Zootown Brew",
      address:"121 W Broadway St",
      city:"Missoula",
      state:"MT"
    ) {
      id
      name
    }
    b7:CreateBusiness(
      id:"b7",
      name:"Hanabi",
      address: "723 California Dr",
      city: "Burlingame",
      state: "CA"
    ) {
      id
      name
    }
     b8:CreateBusiness(
      id:"b8",
      name:"Philz Coffee",
      address: "113 B St",
      city: "San Mateo",
      state: "CA"
    ) {
      id
      name
    }
     b9:CreateBusiness(
      id:"b9",
      name:"Alpha Acid Brewing Company",
      address: "121 Industrial Rd #11",
      city: "Belmont",
      state: "CA"
    ) {
      id
      name
    }
     b10:CreateBusiness(
      id:"b10",
      name:"San Mateo Public Library Central Library",
      address: "55 W 3rd Ave",
      city: "San Mateo",
      state: "CA"
    ) {
      id
      name
    }
    
    c1:CreateCategory(name:"Coffee"){name}
    c2:CreateCategory(name:"Library"){name}
    c3:CreateCategory(name:"Beer"){name}
    c4:CreateCategory(name:"Restaurant"){name}
    c5:CreateCategory(name:"Ramen"){name}
    c6:CreateCategory(name:"Cafe"){name}
    c7:CreateCategory(name:"Deli"){name}
    c8:CreateCategory(name:"Breakfast"){name}
    c9:CreateCategory(name: "Brewery"){name}
    
    
    
    a1: AddBusinessCategory(businessid: "b1", categoryname:"Beer"){id}
    a1a:AddBusinessCategory(businessid: "b1", categoryname:"Brewery"){id}
    a2: AddBusinessCategory(businessid: "b2", categoryname:"Beer"){id}
    a2a:AddBusinessCategory(businessid: "b2", categoryname:"Brewery"){id}
    a3: AddBusinessCategory(businessid: "b3", categoryname:"Restaurant"){id}
    a4: AddBusinessCategory(businessid: "b3", categoryname:"Breakfast"){id}
    a5: AddBusinessCategory(businessid: "b4", categoryname:"Coffee"){id}
    a5a:AddBusinessCategory(businessid: "b4", categoryname:"Restaurant"){id}
    a5b: AddBusinessCategory(businessid: "b4", categoryname:"Cafe"){id}
    a5c: AddBusinessCategory(businessid: "b4", categoryname:"Deli"){id}
    a5d: AddBusinessCategory(businessid: "b4", categoryname:"Breakfast"){id}
    a6: AddBusinessCategory(businessid: "b5", categoryname:"Library"){id}
    a7: AddBusinessCategory(businessid: "b6", categoryname:"Coffee"){id}
    a8: AddBusinessCategory(businessid: "b7", categoryname:"Restaurant"){id}
    a8a: AddBusinessCategory(businessid: "b7", categoryname:"Ramen"){id}
    a9: AddBusinessCategory(businessid: "b8", categoryname:"Coffee"){id}
    a9a: AddBusinessCategory(businessid: "b8", categoryname:"Breakfast"){id}
    a10: AddBusinessCategory(businessid: "b9", categoryname:"Brewery"){id}
    a11:AddBusinessCategory(businessid:"b10", categoryname:"Library"){id}
    
    r1:CreateReview(id:"r1", stars: 4, text: "Great IPA selection!"){id}
    ar1:AddUserReview(userid:"u1",reviewid:"r1"){id}
    ab1:AddReviewBusiness(reviewid:"r1", businessid:"b1"){id}
    
    r2:CreateReview(id:"r2", stars: 5, text: ""){id}
    ar2:AddUserReview(userid:"u3",reviewid:"r2"){id}
    ab2:AddReviewBusiness(reviewid:"r2", businessid:"b1"){id}
    
    r3:CreateReview(id:"r3", stars: 3, text: ""){id}
    ar3:AddUserReview(userid:"u4",reviewid:"r3"){id}
    ab3:AddReviewBusiness(reviewid:"r3", businessid:"b2"){id}
    
    r4:CreateReview(id:"r4", stars: 5, text: ""){id}
    ar4:AddUserReview(userid:"u3",reviewid:"r4"){id}
    ab4:AddReviewBusiness(reviewid:"r4", businessid:"b3"){id}
    
    r5:CreateReview(id:"r5", stars: 4, text: "Best breakfast sandwich at the Farmer's Market. Always get the works."){id}
    ar5:AddUserReview(userid:"u1",reviewid:"r5"){id}
    ab5:AddReviewBusiness(reviewid:"r5", businessid:"b3"){id}
    
    r6:CreateReview(id:"r6", stars: 4, text: ""){id}
    ar6:AddUserReview(userid:"u2",reviewid:"r6"){id}
    ab6:AddReviewBusiness(reviewid:"r6", businessid:"b4"){id}
    
    r7:CreateReview(id:"r7", stars: 3, text: "Not a great selection of books, but fortunately the inter-library loan system is good. Wifi is quite slow. Not many comfortable places to site and read. Looking forward to the new building across the street in 2020!"){id}
    ar7:AddUserReview(userid:"u1",reviewid:"r7"){id}
    ab7:AddReviewBusiness(reviewid:"r7", businessid:"b5"){id}
    
    r8:CreateReview(id:"r8", stars: 5, text: ""){id}
    ar8:AddUserReview(userid:"u4",reviewid:"r8"){id}
    ab8:AddReviewBusiness(reviewid:"r8", businessid:"b6"){id}
    
    r9:CreateReview(id:"r9", stars: 5, text: ""){id}
    ar9:AddUserReview(userid:"u3",reviewid:"r9"){id}
    ab9:AddReviewBusiness(reviewid:"r9", businessid:"b7"){id}
    
    r10:CreateReview(id:"r10", stars: 4, text: ""){id}
    ar10:AddUserReview(userid:"u2",reviewid:"r10"){id}
    ab10:AddReviewBusiness(reviewid:"r10", businessid:""){id}
    
    
  }
`  