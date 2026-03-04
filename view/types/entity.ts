export interface SendMail {
  SMTPId: number;
  recipients: string[];
  subject: string;
  content: string;
}

export interface SendMailApi {
  token: string;
  recipients: string[];
  subject: string;
  content: string;
  files: File[];
}

export interface File {
  name: string;
  data: string;
}