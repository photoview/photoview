export default /* GraphQL */ `
  mutation {
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
      id: "b4"
      name: "Market on Front"
      address: "201 E Front St"
      city: "Missoula"
      state: "MT"
    ) {
      id
      name
    }
    b5: CreateBusiness(
      id: "b5"
      name: "Missoula Public Library"
      address: "301 E Main St"
      city: "Missoula"
      state: "MT"
    ) {
      id
      name
    }
    b6: CreateBusiness(
      id: "b6"
      name: "Zootown Brew"
      address: "121 W Broadway St"
      city: "Missoula"
      state: "MT"
    ) {
      id
      name
    }
    b7: CreateBusiness(
      id: "b7"
      name: "Hanabi"
      address: "723 California Dr"
      city: "Burlingame"
      state: "CA"
    ) {
      id
      name
    }
    b8: CreateBusiness(
      id: "b8"
      name: "Philz Coffee"
      address: "113 B St"
      city: "San Mateo"
      state: "CA"
    ) {
      id
      name
    }
    b9: CreateBusiness(
      id: "b9"
      name: "Alpha Acid Brewing Company"
      address: "121 Industrial Rd #11"
      city: "Belmont"
      state: "CA"
    ) {
      id
      name
    }
    b10: CreateBusiness(
      id: "b10"
      name: "San Mateo Public Library Central Library"
      address: "55 W 3rd Ave"
      city: "San Mateo"
      state: "CA"
    ) {
      id
      name
    }

    c1: CreateCategory(name: "Coffee") {
      name
    }
    c2: CreateCategory(name: "Library") {
      name
    }
    c3: CreateCategory(name: "Beer") {
      name
    }
    c4: CreateCategory(name: "Restaurant") {
      name
    }
    c5: CreateCategory(name: "Ramen") {
      name
    }
    c6: CreateCategory(name: "Cafe") {
      name
    }
    c7: CreateCategory(name: "Deli") {
      name
    }
    c8: CreateCategory(name: "Breakfast") {
      name
    }
    c9: CreateCategory(name: "Brewery") {
      name
    }

    a1: AddBusinessCategories(from: { id: "b1" }, to: { name: "Beer" }) {
      from {
        id
      }
    }
    a1a: AddBusinessCategories(from: { id: "b1" }, to: { name: "Brewery" }) {
      from {
        id
      }
    }
    a2: AddBusinessCategories(from: { id: "b2" }, to: { name: "Beer" }) {
      from {
        id
      }
    }
    a2a: AddBusinessCategories(from: { id: "b2" }, to: { name: "Brewery" }) {
      from {
        id
      }
    }
    a3: AddBusinessCategories(from: { id: "b3" }, to: { name: "Restaurant" }) {
      from {
        id
      }
    }
    a4: AddBusinessCategories(from: { id: "b3" }, to: { name: "Breakfast" }) {
      from {
        id
      }
    }
    a5: AddBusinessCategories(from: { id: "b4" }, to: { name: "Coffee" }) {
      from {
        id
      }
    }
    a5a: AddBusinessCategories(from: { id: "b4" }, to: { name: "Restaurant" }) {
      from {
        id
      }
    }
    a5b: AddBusinessCategories(from: { id: "b4" }, to: { name: "Cafe" }) {
      from {
        id
      }
    }
    a5c: AddBusinessCategories(from: { id: "b4" }, to: { name: "Deli" }) {
      from {
        id
      }
    }
    a5d: AddBusinessCategories(from: { id: "b4" }, to: { name: "Breakfast" }) {
      from {
        id
      }
    }
    a6: AddBusinessCategories(from: { id: "b5" }, to: { name: "Library" }) {
      from {
        id
      }
    }
    a7: AddBusinessCategories(from: { id: "b6" }, to: { name: "Coffee" }) {
      from {
        id
      }
    }
    a8: AddBusinessCategories(from: { id: "b7" }, to: { name: "Restaurant" }) {
      from {
        id
      }
    }
    a8a: AddBusinessCategories(from: { id: "b7" }, to: { name: "Ramen" }) {
      from {
        id
      }
    }
    a9: AddBusinessCategories(from: { id: "b8" }, to: { name: "Coffee" }) {
      from {
        id
      }
    }
    a9a: AddBusinessCategories(from: { id: "b8" }, to: { name: "Breakfast" }) {
      from {
        id
      }
    }
    a10: AddBusinessCategories(from: { id: "b9" }, to: { name: "Brewery" }) {
      from {
        id
      }
    }
    a11: AddBusinessCategories(from: { id: "b10" }, to: { name: "Library" }) {
      from {
        id
      }
    }

    r1: CreateReview(id: "r1", stars: 4, text: "Great IPA selection!", date: { formatted: "2016-01-03"}) {
      id
    }
    ar1: AddUserReviews(from: { id: "u1" }, to: { id: "r1" }) {
      from {
        id
      }
    }
    ab1: AddReviewBusiness(from: { id: "r1" }, to: { id: "b1" }) {
      from {
        id
      }
    }

    r2: CreateReview(id: "r2", stars: 5, text: "", date: { formatted: "2016-07-14"}) {
      id
    }
    ar2: AddUserReviews(from: { id: "u3" }, to: { id: "r2" }) {
      from {
        id
      }
    }
    ab2: AddReviewBusiness(from: { id: "r2" }, to: { id: "b1" }) {
      from {
        id
      }
    }

    r3: CreateReview(id: "r3", stars: 3, text: "", date: { formatted: "2018-09-10"}) {
      id
    }
    ar3: AddUserReviews(from: { id: "u4" }, to: { id: "r3" }) {
      from {
        id
      }
    }
    ab3: AddReviewBusiness(from: { id: "r3" }, to: { id: "b2" }) {
      from {
        id
      }
    }

    r4: CreateReview(id: "r4", stars: 5, text: "", date: { formatted: "2017-11-13"}) {
      id
    }
    ar4: AddUserReviews(from: { id: "u3" }, to: { id: "r4" }) {
      from {
        id
      }
    }
    ab4: AddReviewBusiness(from: { id: "r4" }, to: { id: "b3" }) {
      from {
        id
      }
    }

    r5: CreateReview(
      id: "r5"
      stars: 4
      text: "Best breakfast sandwich at the Farmer's Market. Always get the works."
      date: { formatted: "2018-01-03"}
    ) {
      id
    }
    ar5: AddUserReviews(from: { id: "u1" }, to: { id: "r5" }) {
      from {
        id
      }
    }
    ab5: AddReviewBusiness(from: { id: "r5" }, to: { id: "b3" }) {
      from {
        id
      }
    }

    r6: CreateReview(id: "r6", stars: 4, text: "", date: { formatted: "2018-03-24"}) {
      id
    }
    ar6: AddUserReviews(from: { id: "u2" }, to: { id: "r6" }) {
      from {
        id
      }
    }
    ab6: AddReviewBusiness(from: { id: "r6" }, to: { id: "b4" }) {
      from {
        id
      }
    }

    r7: CreateReview(
      id: "r7"
      stars: 3
      text: "Not a great selection of books, but fortunately the inter-library loan system is good. Wifi is quite slow. Not many comfortable places to site and read. Looking forward to the new building across the street in 2020!"
      date: { formatted: "2015-08-29"}
    ) {
      id
    }
    ar7: AddUserReviews(from: { id: "u1" }, to: { id: "r7" }) {
      from {
        id
      }
    }
    ab7: AddReviewBusiness(from: { id: "r7" }, to: { id: "b5" }) {
      from {
        id
      }
    }

    r8: CreateReview(id: "r8", stars: 5, text: "", date: { formatted: "2018-08-11"}) {
      id
    }
    ar8: AddUserReviews(from: { id: "u4" }, to: { id: "r8" }) {
      from {
        id
      }
    }
    ab8: AddReviewBusiness(from: { id: "r8" }, to: { id: "b6" }) {
      from {
        id
      }
    }

    r9: CreateReview(id: "r9", stars: 5, text: "", date: { formatted: "2016-11-21"}) {
      id
    }
    ar9: AddUserReviews(from: { id: "u3" }, to: { id: "r9" }) {
      from {
        id
      }
    }
    ab9: AddReviewBusiness(from: { id: "r9" }, to: { id: "b7" }) {
      from {
        id
      }
    }

    r10: CreateReview(id: "r10", stars: 4, text: "", date: { formatted: "2015-12-15"}) {
      id
    }
    ar10: AddUserReviews(from: { id: "u2" }, to: { id: "r10" }) {
      from {
        id
      }
    }
    ab10: AddReviewBusiness(from: { id: "r10" }, to: { id: "b2" }) {
      from {
        id
      }
    }
  }
`;
