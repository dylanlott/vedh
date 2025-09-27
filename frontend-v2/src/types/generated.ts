// Temporary hand-written GraphQL types until codegen is added.
export interface LoginMutation {
  login: {
    ID: string;
    Username: string;
    Token: string;
  };
}

export interface LoginMutationVariables {
  username: string;
  password: string;
}

export interface SignupMutation {
  signup: {
    ID: string;
    Username: string;
    Token: string;
  };
}

export interface SignupMutationVariables {
  username: string;
  password: string;
}
