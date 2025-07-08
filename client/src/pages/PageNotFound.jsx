import React from "react";
import { Link } from "react-router-dom";

const PageNotFound = () => {
  return (
    <div>
      <main className="h-screen w-full flex flex-col justify-center items-center bg-accent">
        <h1 className="text-9xl font-extrabold text-primary tracking-widest">404</h1>
        <div className="bg-secondary px-2 text-sm rounded rotate-12 absolute">
          <span className="text-secondary-foreground">Page Not Found</span>
        </div>
        <button className="mt-5">
          <a className="relative inline-block text-sm font-medium text-secondary group active:text-secondary focus:outline-none focus:ring">
            <span className="absolute inset-0 transition-transform translate-x-0.5 translate-y-0.5 bg-primary group-hover:translate-y-0 group-hover:translate-x-0"></span>

            <span className="relative block px-8 py-3 bg-primary border border-current">
              <Link to="/dashboard">Go Home</Link>
            </span>
          </a>
        </button>
      </main>
    </div>
  );
};

export default PageNotFound;
